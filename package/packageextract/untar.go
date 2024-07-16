package packageextract

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func UnTarGZ(in, out string) error {
	if _, err := os.Stat(in); err != nil {
		return errors.Wrapf(err, "get file %s's stat error", in)
	}
	fr, err := os.Open(in)
	if err != nil {
		return errors.Wrapf(err, "open file %s error", in)
	}
	defer fr.Close()
	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case hdr == nil:
			continue
		}

		dst := filepath.Join(out, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				if err := os.MkdirAll(dst, hdr.FileInfo().Mode().Perm()); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, hdr.FileInfo().Mode().Perm())
			if err != nil {
				return err
			}
			_, err = io.Copy(f, tr)
			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := os.Symlink(hdr.Linkname, dst); err != nil {
				return err
			}
		case tar.TypeLink:
			if err := os.Link(hdr.Linkname, dst); err != nil {
				return err
			}
		}
	}
}
