package logs

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"log/slog"
	"os"
	"time"
)

var (
	ErrCorruptedLine = errors.New("corrupted log line")
	ErrNotExists     = errors.New("key doesn't exists")
	ErrClosedCh      = errors.New("channel closed or full")
)

const (
	syncInterval = time.Millisecond * 10
	compInterval = time.Minute * 30
	writesChSize = 32
	opChSize     = 32
	headerLen    = 4
	newLine      = "\n"
	tmpName      = "compaction-*"
)

type Log struct {
	writesCh chan []byte
	closedCh chan struct{}
	file     *os.File
}

func New(fileName string, mem map[string][]byte) (*Log, error) {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	log := &Log{
		writesCh: make(chan []byte, writesChSize),
		closedCh: make(chan struct{}, 1),
		file:     f,
	}
	if err := log.run(mem); err != nil {
		return nil, err
	}
	return log, nil
}

func (l *Log) run(mem map[string][]byte) error {
	n, err := read(l.file, mem)
	if err != nil {
		if !errors.Is(err, ErrCorruptedLine) {
			return err
		}
		if err := l.file.Truncate(int64(n)); err != nil {
			return err
		}
	}
	go l.listen()
	return nil
}

func (l *Log) listen() {
	syncTicker := time.NewTicker(syncInterval)
	compactTicker := time.NewTicker(compInterval)
	defer func() {
		syncTicker.Stop()
		compactTicker.Stop()
		l.close()
	}()
loop:
	for {
		select {
		case <-syncTicker.C:
			if err := l.sync(); err != nil {
				slog.Error("log/sync", "err", err)
				break loop
			}
		case <-compactTicker.C:
			if err := l.compact(); err != nil {
				slog.Error("log/compaction", "err", err)
				break loop
			}
		case <-l.closedCh:
			break loop
		}
	}
}

// read from the given readseeker
// from the first line to the last
func read(r io.ReadSeeker, mem map[string][]byte) (int, error) {
	_, err := r.Seek(0, 0)
	if err != nil {
		return 0, err
	}
	scanner := bufio.NewScanner(r)
	readN := 0
	for scanner.Scan() {
		n, err := readLine(scanner.Bytes(), mem)
		if err != nil {
			return readN, err
		}
		readN += n + len(newLine)
	}
	return readN, nil
}

func (l *Log) Append(data []byte, keyLen int) error {
	line := formatLine(data, keyLen)
	select {
	case l.writesCh <- line:
	default:
		return ErrClosedCh
	}
	return nil
}

func (l *Log) compact() error {
	if err := l.sync(); err != nil {
		return err
	}
	tmp, err := l.copyFile()
	if err != nil {
		slog.Error("compact", "err", err)
		slog.Info("cleaning tmp", "path", tmp.Name())
		if err := l.cleanTmp(tmp); err != nil {
			return err
		}
		slog.Info("cleaning tmp successful")
		return err
	}
	return l.renameTmp(tmpName)
}

func (l *Log) copyFile() (*os.File, error) {
	tmp, err := os.CreateTemp(".", tmpName)
	if err != nil {
		return nil, err
	}

	mem := make(map[string][]byte)

	_, err = read(l.file, mem)
	if err != nil {
		if !errors.Is(err, ErrCorruptedLine) {
			return nil, err
		}
		slog.Error("log/read", "err", err)
	}
	buf := bytes.NewBuffer(make([]byte, 1024))
	for key, val := range mem {
		buf.WriteString(key)
		buf.Write(val)
		line := formatLine(buf.Bytes(), len(key))
		_, err := tmp.Write(line)
		if err != nil {
			return nil, err
		}
		buf.Reset()
	}
	if err := tmp.Sync(); err != nil {
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}
	return tmp, nil
}

func (l *Log) cleanTmp(tmp *os.File) error {
	tmp.Close()
	if err := os.Remove(tmp.Name()); err != nil {
		return err
	}
	return nil
}

func (l *Log) renameTmp(tmpName string) error {
	name := l.file.Name()
	if err := l.file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, name); err != nil {
		return err
	}
	f, err := os.OpenFile(l.file.Name(), os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	l.file = f
	return nil
}

func (l *Log) sync() error {
loop:
	for {
		select {
		case w := <-l.writesCh:
			_, err := l.file.Write(w)
			if err != nil {
				return err
			}
		default:
			break loop
		}
	}
	return l.file.Sync()
}

func (l *Log) Close() error {
	select {
	case l.closedCh <- struct{}{}:
		return nil
	default:
		return ErrClosedCh
	}
}

func (l *Log) close() {
	slog.Info("closing log...")
	if err := l.sync(); err != nil {
		slog.Error("log/sync", "err", err)
	}
	if err := l.file.Close(); err != nil {
		slog.Error("log/file", "err", err)
	}
}

func readLine(data []byte, mem map[string][]byte) (int, error) {
	length := len(data)
	sum := crc32.ChecksumIEEE(data[:length-headerLen])
	expectedSum := binary.LittleEndian.Uint32(data[length-headerLen:])
	if sum != expectedSum {
		return 0, ErrCorruptedLine
	}
	keyLen := binary.LittleEndian.Uint32(data[:headerLen])
	body := data[headerLen : length-headerLen]
	if int(keyLen) == len(body) {
		delete(mem, string(body))
	} else {
		mem[string(body[:int(keyLen)])] = body[keyLen:]
	}
	return length, nil
}

func formatLine(data []byte, keyLen int) []byte {
	header := binary.LittleEndian.AppendUint32(nil, uint32(keyLen))

	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, header)
	buffer.Write(data)

	sum := crc32.ChecksumIEEE(buffer.Bytes())

	binary.Write(buffer, binary.LittleEndian, sum)
	buffer.WriteString(newLine)
	return buffer.Bytes()
}
