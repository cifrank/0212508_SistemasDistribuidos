package log

import (
	"fmt"
	"io"
	"os"
	"path"

	api "github.com/cifrank/0212508_SistemasDistribuidos/api/v1"
	"google.golang.org/protobuf/proto"
)

type segment struct {
	store                  *store
	index                  *index
	baseOffset, nextOffset uint64
	config                 Config
}

// Copiado del Notion gracias a la misericordia del prof
func newSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}
	var err error
	storeFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".store")),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}
	if s.store, err = newStore(storeFile); err != nil {
		return nil, err
	}
	indexFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index")),
		os.O_RDWR|os.O_CREATE,
		0644,
	)
	if err != nil {
		return nil, err
	}
	if s.index, err = newIndex(indexFile, c); err != nil {
		return nil, err
	}
	if off, _, err := s.index.Read(-1); err != nil {
		s.nextOffset = baseOffset
	} else {
		s.nextOffset = baseOffset + uint64(off) + 1
	}

	return s, nil
}

func (s *segment) Append(record *api.Record) (off uint64, err error) {
	off = s.nextOffset
	// Arregle errores, ahora chequea si el segmento está lleno antes de escribir en vez de al final, igual, retorna un 0 mejor en vez de la posicion actual (que la habia puesto para indicar en donde quedo)
	if s.IsMaxed() {
		return 0, io.EOF
	}
	data, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}

	_, pos, err := s.store.Append(data)
	if err != nil {
		return 0, err
	}
	// escribe en el índice la posición del registro con respecto al offset base
	if err = s.index.Write(
		uint32(s.nextOffset-uint64(s.baseOffset)),
		pos,
	); err != nil {
		return 0, err
	}

	s.nextOffset++

	return off, nil
}

// Lee y decodifica un registro en la posición del offset dado
func (s *segment) Read(off uint64) (*api.Record, error) {
	_, pos, err := s.index.Read(int64(off - s.baseOffset))
	if err != nil {
		return nil, err
	}

	b, err := s.store.Read(pos)
	if err != nil {
		return nil, err
	}

	record := &api.Record{}
	err = proto.Unmarshal(b, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s *segment) IsMaxed() bool {
	return s.store.size >= s.config.Segment.MaxStoreBytes || s.index.size >= s.config.Segment.MaxIndexBytes
}

func (s *segment) Remove() error {
	if err := s.index.Close(); err != nil {
		return err
	}
	if err := s.store.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.index.file.Name()); err != nil {
		return err
	}
	if err := os.Remove(s.store.File.Name()); err != nil {
		return err
	}
	return nil
}

func (s *segment) Close() error {
	if err := s.index.Close(); err != nil {
		return err
	}
	if err := s.store.Close(); err != nil {
		return err
	}
	return nil
}
