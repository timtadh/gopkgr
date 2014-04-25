package tar

import (
	"archive/tar"
	"compress/gzip"
	pathpkg "path"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func Empty(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return true
	}
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return true
	}
	return len(dir) == 0
}

func dirpath(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}

func Process(prefix, path string, archive *tar.Writer) error {
	abs_path := pathpkg.Join(prefix, path)
	stat, err := os.Stat(abs_path)
	if err != nil { panic(err) }
	if stat.IsDir() {
		return ProcessDir(prefix, path, archive)
	} else {
		return ProcessFile(prefix, path, archive)
	}
}

func ProcessFile(prefix, path string, archive *tar.Writer) error {
	name := pathpkg.Base(path)
	if strings.HasPrefix(name, ".") {
		return nil
	}
	abs_path := pathpkg.Join(prefix, path)
	stat, err := os.Stat(abs_path)
	if err != nil {
		return err
	}
	f, err := os.Open(abs_path)
	if err != nil {
		return err
	}
	hdr, err := tar.FileInfoHeader(stat, "")
	if err != nil {
		return err
	}
	hdr.Name = path
	if err := archive.WriteHeader(hdr); err != nil {
		log.Fatalln(err)
	}
	chunk := make([]byte, 4096)
	for {
		n, err := f.Read(chunk)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		for m := 0; m < n; {
			s, err := archive.Write(chunk[m:n])
			if err != nil {
				return err
			} else if s == 0 {
				return fmt.Errorf("Wrote 0 bytes")
			}
			m += s
		}
	}
	return nil
}

func ProcessDir(prefix, path string, archive *tar.Writer) error {
	abs_path := pathpkg.Join(prefix, path)
	stat, err := os.Stat(abs_path)
	if err != nil {
		return err
	}
	dir, err := ioutil.ReadDir(abs_path)
	if err != nil {
		return err
	}
	hdr, err := tar.FileInfoHeader(stat, "")
	if err != nil {
		return err
	}
	hdr.Name = dirpath(path)
	if err := archive.WriteHeader(hdr); err != nil {
		log.Fatalln(err)
	}
	for _, info := range dir {
		err := Process(prefix, pathpkg.Join(path, info.Name()), archive)
		if err != nil {
			return err
		}
	}
	return nil
}

func Archive(prefix, path, target string) error {
	if !Exists(pathpkg.Join(prefix, path)) {
		return fmt.Errorf("Source directory did not exist")
	}
	if Exists(target) {
		return fmt.Errorf("Cowardly refusing to over-write existing target")
	}
	cleanup := func(f *os.File) {
		f.Close()
		os.Remove(target)
	}
	f, err := os.Create(target)
	if err != nil {
		log.Fatalln(err)
	}
	g := gzip.NewWriter(f)
	t := tar.NewWriter(g)
	if err := Process(prefix, path, t); err != nil {
		cleanup(f)
		return err
	}
	if err := t.Close(); err != nil {
		cleanup(f)
		return err
	}
	if err := g.Close(); err != nil {
		cleanup(f)
		return err
	}
	return f.Close()
}

func open(prefix, source string) (*tar.Reader, func(), error) {
	if !Exists(source) {
		return nil, nil, fmt.Errorf("source archive does not exist!")
	}
	if err := os.MkdirAll(prefix, 0775); err != nil {
		return nil, nil, err
	}
	f, err := os.Open(source)
	if err != nil {
		return nil, nil, err
	}
	g, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	t := tar.NewReader(g)
	return t, func() {
		g.Close()
		f.Close()
	}, nil
}

func Unpack(prefix, source string) error {
	t, closer, err := open(prefix, source)
	if err != nil {
		return err
	}
	for {
		hdr, err := t.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			closer()
			return err
		}
		abs_path := pathpkg.Join(prefix, hdr.Name)
		if !hdr.FileInfo().IsDir() {
			if Exists(abs_path) {
				closer()
				return fmt.Errorf("Cowardly refusing to over-write %s", abs_path)
			}
		}
	}
	closer()

	t, closer, err = open(prefix, source)
	if err != nil {
		return err
	}
	defer closer()
	for {
		hdr, err := t.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		abs_path := pathpkg.Join(prefix, hdr.Name)
		if hdr.FileInfo().IsDir() {
			if err := os.Mkdir(abs_path, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		} else {
			f, err := os.Create(abs_path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, t); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func Remove(prefix, source string) error {
	t, closer, err := open(prefix, source)
	if err != nil {
		return err
	}
	files := make([]string, 0)
	for {
		hdr, err := t.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			closer()
			return err
		}
		abs_path := pathpkg.Join(prefix, hdr.Name)
		files = append(files, abs_path)
	}
	closer()
	for i := len(files)-1; i >= 0; i-- {
		path := files[i]
		stat, err := os.Stat(path)
		if err != nil {
			continue
		}
		if stat.IsDir() {
			if Empty(path) {
				os.Remove(path)
			}
		} else {
			os.Remove(path)
		}
	}
	return nil
}
