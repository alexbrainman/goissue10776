package main

import (
	"debug/pe"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"unsafe"
)

const (
	IMAGE_SYM_CLASS_FILE = 103
)

func printSection(path string, name string) (*pe.Section, error) {
	f, err := pe.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sect := f.Section(name)
	if sect == nil {
		return nil, fmt.Errorf("could not find section %q", name)
	}
	data, err := sect.Data()
	if err != nil {
		return nil, err
	}
	fmt.Print(hex.Dump(data))
	s := *sect
	return &s, nil
}

type Reloc struct {
	VirtualAddress   uint32
	SymbolTableIndex uint32
	Type             int16
}

func printRelocations(path string, sect *pe.Section) error {
	fmt.Println()
	if sect.NumberOfRelocations <= 0 {
		fmt.Println("no relocations")
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Seek(int64(sect.PointerToRelocations), os.SEEK_SET); err != nil {
		return err
	}
	relocs := make([]Reloc, sect.NumberOfRelocations)
	if err := binary.Read(f, binary.LittleEndian, relocs); err != nil {
		return err
	}

	fmt.Println("Relocations:")
	fmt.Println()
	fmt.Printf("idx  type   address symtab_index\n")
	fmt.Printf("--- ----- --------- ------------\n")
	for i, r := range relocs {
		fmt.Printf("%3d %5x %9x %12d\n", i, r.Type, r.VirtualAddress, r.SymbolTableIndex)
	}
	return nil
}

func dumpSection(path string, name string) error {
	sect, err := printSection(path, name)
	if err != nil {
		return err
	}
	return printRelocations(path, sect)
}

func printAlignments(f *pe.File) {
	fmt.Println()
	switch oh := f.OptionalHeader.(type) {
	case nil:
		fmt.Printf("no section or file alignment (no optional header present)\n")
	case *pe.OptionalHeader32:
		fmt.Printf("section alignment is 0x%x\n", oh.SectionAlignment)
		fmt.Printf("file alignment is 0x%x\n", oh.FileAlignment)
	case *pe.OptionalHeader64:
		fmt.Printf("section alignment is 0x%x\n", oh.SectionAlignment)
		fmt.Printf("file alignment is 0x%x\n", oh.FileAlignment)
	default:
		panic("unknown OptionalHeader type")
	}
}

func printSections(path string) (*pe.FileHeader, error) {
	f, err := pe.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fmt.Println()
	fmt.Println("Sections:")
	fmt.Println()
	fmt.Printf("idx virtual virtual    disk    disk   reloc  reloc     mask\n")
	fmt.Printf("    address    size  offset    size  offset    qty         \n")
	fmt.Printf("--- ------- ------- ------- ------- ------- ------ --------\n")
	for i, s := range f.Sections {
		fmt.Printf("%3d %7x %7x %7x %7x %7x %6d %x %s\n",
			i+1, s.VirtualAddress, s.VirtualSize, s.Offset, s.Size,
			s.PointerToRelocations, s.NumberOfRelocations,
			s.Characteristics, s.Name)
	}
	printAlignments(f)
	fh := f.FileHeader
	return &fh, nil
}

func cstring(b []byte) string {
	var i int
	for i = 0; i < len(b) && b[i] != 0; i++ {
	}
	return string(b[0:i])
}

// getString extracts a string from symbol string table.
func getString(section []byte, start int) (string, bool) {
	if start < 0 || start >= len(section) {
		return "", false
	}

	for end := start; end < len(section); end++ {
		if section[end] == 0 {
			return string(section[start:end]), true
		}
	}
	return "", false
}

func printSymbols(f *os.File, fh *pe.FileHeader) error {
	fmt.Println()
	if fh.NumberOfSymbols <= 0 {
		fmt.Println("no symbols")
		return nil
	}

	// Get COFF string table, which is located at the end of the COFF symbol table.
	if _, err := f.Seek(int64(fh.PointerToSymbolTable+pe.COFFSymbolSize*fh.NumberOfSymbols), os.SEEK_SET); err != nil {
		return err
	}
	var l uint32
	if err := binary.Read(f, binary.LittleEndian, &l); err != nil {
		return err
	}
	ss := make([]byte, l)
	if _, err := f.ReadAt(ss, int64(fh.PointerToSymbolTable+pe.COFFSymbolSize*fh.NumberOfSymbols)); err != nil {
		return err
	}

	// Process COFF symbol table.
	if _, err := f.Seek(int64(fh.PointerToSymbolTable), os.SEEK_SET); err != nil {
		return err
	}
	syms := make([]pe.COFFSymbol, fh.NumberOfSymbols)
	if err := binary.Read(f, binary.LittleEndian, syms); err != nil {
		return err
	}

	fmt.Println("Symbols:")
	fmt.Println()
	fmt.Printf("idx  type section     value  class   aux name\n")
	fmt.Printf("--- ----- ------- --------- ------ ----- ----------------\n")
	aux := uint8(0)
	auxstart := 0
	for i, s := range syms {
		var name string
		if s.Name[0] == 0 && s.Name[1] == 0 && s.Name[2] == 0 && s.Name[3] == 0 {
			if s.Name[4] == 0 && s.Name[5] == 0 && s.Name[6] == 0 && s.Name[7] == 0 {
				name = ""
			} else {
				si := int(binary.LittleEndian.Uint32(s.Name[4:]))
				name, _ = getString(ss, si)
			}
		} else {
			name = cstring(s.Name[:])
		}
		if aux > 0 {
			aux--
			auxdata := ((*[18]byte)(unsafe.Pointer(&s)))[:]
			switch {
			case syms[auxstart].StorageClass == IMAGE_SYM_CLASS_FILE:
				fmt.Printf("    %s\n", cstring(auxdata))
			default:
				fmt.Printf("    %s\n", hex.EncodeToString(auxdata))
			}
		} else {
			aux = s.NumberOfAuxSymbols
			auxstart = i
			fmt.Printf("%3d %5d %7d %9x %6d %5d %v %s\n",
				i, s.Type, s.SectionNumber, s.Value, s.StorageClass,
				s.NumberOfAuxSymbols, hex.EncodeToString(s.Name[:]), name)
		}
	}

	// TODO: I am not sure I want to print coff strings here
	return nil
	fmt.Println()
	fmt.Println("Strings:")
	fmt.Println()
	fmt.Print(hex.Dump(ss))
	return nil
}

func printExtraAfterSectionsList(path string, fh *pe.FileHeader) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	printSymbols(f, fh)
	return nil
}

func listSections(path string) error {
	fh, err := printSections(path)
	if err != nil {
		return err
	}
	return printExtraAfterSectionsList(path, fh)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage is: %s <exe-name> [<section-name>]\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	switch len(flag.Args()) {
	case 1:
		err := listSections(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
	case 2:
		err := dumpSection(flag.Arg(0), flag.Arg(1))
		if err != nil {
			log.Fatal(err)
		}
	default:
		flag.Usage()
	}
}
