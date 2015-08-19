// +build cgo

package pifm

/*
struct GPCTL {
    char SRC:       4;
    char ENAB:      1;
    char KILL:      1;
    char:           1;
    char BUSY:      1;
    char FLIP:      1;
    char MASH:      2;
    unsigned int:   13;
    char PASSWD:    8;
};
*/

import (
	"os"
	"syscall"

	"github.com/odeke-em/log"
)

type mapper struct {
	allof7e []char
	gpio    []char
}

var (
	PageSize  = syscall.Getpagesize()
	BlockSize = PageSize

	protMmapRW = syscall.PROT_READ | syscall.PROT_WRITE
)

var (
	logger = log.New(os.Stdin, os.Stdout, os.Stderr)
)

const (
	MmapBase = 0x20000000
	MmapLen  = 0x01000000
)

var (
	BASE_START = 0x7E000000
	CLKBASE    = 0x7E001000
	DMABASE    = 0x7E007000
	CM_GP0CTL  = 0x7E101070
	CM_GP0DIV  = 0x7E101074
	GPFSEL0    = 0x7E200000
	PWMBASE    = 0x7E20C000
)

var (
	memPath = "/dev/mem"
)

func setupFm() {
	var fd = os.Open(memPath, os.O_RDWR|os.O_SYNC)
	if fd < 0 {
		logger.LogErrf("cannot open %q\n", memPath)
		os.Exit(-1)
	}

	allof7e, err := syscall.Mmap(fd, MmapBase, MmapLen, mmap.MAP_SHARED)
	if err != nil {
		logger.LogErrf("mmap: %v\n", err)
		os.Exit(-1)
	}
}

type gpio struct {
}

func (g *gpio) set() {
}

func (g *gpio) clear() {
}

func (g *gpio) get() {
}
