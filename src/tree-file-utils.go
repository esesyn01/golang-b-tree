package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"os"
	"reflect"
)

func get_file(name string) *os.File {
	file, err := os.Open(name)
	if err != nil {
		log.Fatal("File cannot be opened. Aborting...")
	}
	return file
}

func write_page_to_file(page tree_page, filename string, offset int) {
	file := get_file(filename)
	off := int64(offset)
	_, err := file.Seek(off, 0)
	if err != nil {
		file.Close()
		log.Fatalln("Cannot achieve desired offset. Aborting...")
	}
	err = binary.Write(file, binary.LittleEndian, page)
	if err != nil {
		file.Close()
		log.Fatalln("Write failed. Aborting...")
	}
	t := reflect.TypeOf(page)
	size := t.Size()
	if size != TREE_PAGE_SIZE {
		buf := bytes.NewBuffer(make([]byte, TREE_PAGE_SIZE-size))
		err = binary.Write(file, binary.LittleEndian, buf)
		if err != nil {
			log.Fatalln("Cannot add blank zeroes. Aborting...")
		}
	}
	file.Close()
}

func read_page_from_file(filename string, offset int) tree_page {
	file := get_file(filename)
	buffer := make([]byte, TREE_PAGE_SIZE)
	_, err := file.ReadAt(buffer, int64(offset))
	{
		if err != nil {
			if err != io.EOF {
				log.Fatalln(err)
			}
		}
	}
	var page tree_page
	buffer1 := bytes.NewBuffer(buffer[0:HEADER_SIZE])
	buffer2 := bytes.NewBuffer(buffer[HEADER_SIZE : HEADER_SIZE+4])
	err = binary.Read(buffer1, binary.LittleEndian, &page.header)
	if err != nil {
		log.Fatalln("Cannot read header. Aborting...")
	}
	err = binary.Read(buffer2, binary.LittleEndian, &page.first_child_offset)
	if err != nil {
		log.Fatalln("Cannot read first pointer. Aborting...")
	}
	i := 0
	for int32(i) < page.header.records_number {
		buffer_record := bytes.NewBuffer(buffer[HEADER_SIZE+4+i*SINGLE_TREE_RECORD_SIZE : HEADER_SIZE+4+(i+1)*SINGLE_TREE_RECORD_SIZE])
		var temp_tree_record tree_record
		err = binary.Read(buffer_record, binary.LittleEndian, &temp_tree_record)
		if err != nil {
			log.Fatalln("Cannot get record from file. Aborting...")
		}
		page.records = append(page.records, temp_tree_record)
		i += 1
	}
	file.Close()
	return page
}
