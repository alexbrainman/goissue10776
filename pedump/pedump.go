package main

import (
	"debug/pe"
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

func printRelocations(f *pe.File, sect *pe.Section) error {
	fmt.Println()
	if sect.NumberOfRelocations <= 0 {
		fmt.Println("no relocations")
		return nil
	}

	fmt.Println("Relocations:")
	fmt.Println()
	fmt.Printf("idx  type   address symtab_index\n")
	fmt.Printf("--- ----- --------- ------------\n")
	for i, r := range sect.Relocs {
		fmt.Printf("%3d %5x %9x %12d\n", i, r.Type, r.VirtualAddress, r.SymbolTableIndex)
	}
	return nil
}

func dumpSection(path string, name string) error {
	f, err := pe.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sect := f.Section(name)
	if sect == nil {
		return fmt.Errorf("could not find section %q", name)
	}
	data, err := sect.Data()
	if err != nil {
		return err
	}
	fmt.Print(hex.Dump(data))

	return printRelocations(f, sect)
}

func cstring(b []byte) string {
	var i int
	for i = 0; i < len(b) && b[i] != 0; i++ {
	}
	return string(b[0:i])
}

func printSymbols(f *pe.File) error {
	fmt.Println()
	if f.FileHeader.NumberOfSymbols <= 0 {
		fmt.Println("no symbols")
		return nil
	}
	fmt.Println("Symbols:")
	fmt.Println()
	fmt.Printf("idx  type section     value  class   aux name\n")
	fmt.Printf("--- ----- ------- --------- ------ ----- ----------------\n")
	aux := uint8(0)
	auxstart := 0
	for i, s := range f.COFFSymbols {
		if aux > 0 {
			aux--
			auxdata := ((*[18]byte)(unsafe.Pointer(&s)))[:]
			switch {
			case f.COFFSymbols[auxstart].StorageClass == IMAGE_SYM_CLASS_FILE:
				fmt.Printf("    %s\n", cstring(auxdata))
			default:
				fmt.Printf("    %s\n", hex.EncodeToString(auxdata))
			}
		} else {
			name, err := s.FullName(f.StringTable)
			if err != nil {
				return fmt.Errorf("failed to read full name for symbol=%v: %v", s.Name[:], err)
			}
			aux = s.NumberOfAuxSymbols
			auxstart = i
			fmt.Printf("%3d %5d %7d %9x %6d %5d %v %s\n",
				i, s.Type, s.SectionNumber, s.Value, s.StorageClass,
				s.NumberOfAuxSymbols, hex.EncodeToString(s.Name[:]), name)
		}
	}
	return nil
}

func listSections(path string) error {
	f, err := pe.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// print sections
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

	// print alignments
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

	// print symbols
	return printSymbols(f)

	// TODO: I am not sure I want to print coff strings here
	fmt.Println()
	fmt.Println("Strings:")
	fmt.Println()
	fmt.Print(hex.Dump(f.StringTable))
	return nil
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
