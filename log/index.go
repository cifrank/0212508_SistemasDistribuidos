package log

import (
	"fmt"
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

func newIndex(f *os.File, c Config) (*index, error) {
	fi, err := f.Stat()
	if err != nil {
		err = fmt.Errorf("f.Stat: %w", err)
		return nil, err
	}
	size := uint64(fi.Size())
	// trunca el archivo a un tamaño máximo ya que me estaba dando errores al hacer el test y parece arreglarlo esto
	f.Truncate(int64(c.Segment.MaxIndexBytes))
	// lo mapea en memoria con permisos de lectura y escritura
	mmap, err := gommap.Map(f.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED)
	if err != nil {
		err = fmt.Errorf("gommap.Map: %w", err)
		return nil, err
	}
	i := &index{
		file: f,
		mmap: mmap,
		size: size,
	}
	if size == 0 {
		enc.PutUint32(i.mmap, 0)
	}
	return i, nil
}

// sincroniza los cambios en memoria con el archivo para poder cerralo sin errores (idealmente)
func (i *index) Close() error {
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	if err := i.file.Sync(); err != nil {
		return err
	}
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	if err := i.file.Close(); err != nil {
		return err
	}
	return nil
}

func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}
	pos = uint64(out) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

// Agrega entradas mientras haya espacio en el archivo, aka el mapa de memoria
func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += entWidth
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
