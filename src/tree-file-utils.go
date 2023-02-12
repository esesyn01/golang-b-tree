package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

func get_file(name string) *os.File {
	file, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal("File cannot be opened. Aborting...")
	}
	return file
}

func create_bin_file(name string) {
	file, err := os.Create(name)
	if err != nil {
		log.Fatal("File cannot be created. Aborting...")
	}
	file.Close()
}

func remove_file(name string) {
	err := os.Remove(name)
	if err != nil {
		log.Fatal("File cannot be erased. Aborting...")
	}
}

func get_offset_for_new_node() int32 {
	if len(free_list_pages) > 0 {
		addr := free_list_pages[len(free_list_pages)-1]
		free_list_pages = free_list_pages[:len(free_list_pages)-1]
		return addr
	} else {
		file := get_file(TREE_FILE_NAME)
		fileinfo, _ := file.Stat()
		size := fileinfo.Size()
		file.Close()
		return int32(size)
	}
}

func write_page_to_file(page tree_page) {
	file := get_file(TREE_FILE_NAME)
	off := int64(page.header.own_offset)
	_, err := file.Seek(off, 0)
	if err != nil {
		file.Close()
		log.Fatalln("Cannot achieve desired offset. Aborting...")
	}
	err = binary.Write(file, binary.LittleEndian, page.header)
	if err != nil {
		file.Close()
		fmt.Println(err)
		log.Fatalln("Write failed. Aborting...")
	}
	_ = binary.Write(file, binary.LittleEndian, page.first_child_offset)
	i := 0
	bytes_written := HEADER_SIZE + INT_SIZE
	for int32(i) < page.header.records_number {
		err = binary.Write(file, binary.LittleEndian, page.records[i])
		if err != nil {
			file.Close()
			fmt.Println(err)
			log.Fatalln("Write failed. Aborting...")
		}
		bytes_written += SINGLE_TREE_RECORD_SIZE
		i += 1
	}
	if bytes_written != TREE_PAGE_SIZE {
		buf := bytes.NewBuffer(make([]byte, TREE_PAGE_SIZE-bytes_written))
		err = binary.Write(file, binary.LittleEndian, buf.Bytes())
		if err != nil {
			fmt.Println(err)
			log.Fatalln("Cannot add blank zeroes. Aborting...")
		}
	}
	file.Close()
}

func read_page_from_file(offset int32) tree_page {
	file := get_file(TREE_FILE_NAME)
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
	var header tree_page_header
	buffer1 := bytes.NewBuffer(buffer[0:INT_SIZE])
	buffer2 := bytes.NewBuffer(buffer[INT_SIZE : INT_SIZE*2])
	buffer3 := bytes.NewBuffer(buffer[INT_SIZE*2 : INT_SIZE*3])
	buffer20 := bytes.NewBuffer(buffer[HEADER_SIZE : HEADER_SIZE+INT_SIZE])
	err = binary.Read(buffer1, binary.LittleEndian, &header.parent_offset)
	if err != nil {
		fmt.Println(err)
		log.Fatalln("Cannot read header. Aborting...")
	}
	err = binary.Read(buffer2, binary.LittleEndian, &header.records_number)
	if err != nil {
		fmt.Println(err)
		log.Fatalln("Cannot read header. Aborting...")
	}
	err = binary.Read(buffer3, binary.LittleEndian, &header.own_offset)
	if err != nil {
		fmt.Println(err)
		log.Fatalln("Cannot read header. Aborting...")
	}
	page.header = header
	err = binary.Read(buffer20, binary.LittleEndian, &page.first_child_offset)
	if err != nil {
		log.Fatalln("Cannot read first pointer. Aborting...")
	}
	i := 0
	for int32(i) < page.header.records_number {
		buffer_record1 := bytes.NewBuffer(buffer[HEADER_SIZE+INT_SIZE+i*SINGLE_TREE_RECORD_SIZE : HEADER_SIZE+INT_SIZE+i*SINGLE_TREE_RECORD_SIZE+INT_SIZE])
		buffer_record2 := bytes.NewBuffer(buffer[HEADER_SIZE+INT_SIZE+i*SINGLE_TREE_RECORD_SIZE+INT_SIZE : HEADER_SIZE+INT_SIZE+i*SINGLE_TREE_RECORD_SIZE+INT_SIZE*2])
		buffer_record3 := bytes.NewBuffer(buffer[HEADER_SIZE+INT_SIZE+i*SINGLE_TREE_RECORD_SIZE+INT_SIZE*2 : HEADER_SIZE+INT_SIZE+i*SINGLE_TREE_RECORD_SIZE+INT_SIZE*3])
		var temp_tree_record tree_record
		err = binary.Read(buffer_record1, binary.LittleEndian, &temp_tree_record.key)
		if err != nil {
			log.Fatalln("Cannot get record from file. Aborting...")
		}
		err = binary.Read(buffer_record2, binary.LittleEndian, &temp_tree_record.record_offset)
		if err != nil {
			log.Fatalln("Cannot get record from file. Aborting...")
		}
		err = binary.Read(buffer_record3, binary.LittleEndian, &temp_tree_record.child_page_offset)
		if err != nil {
			log.Fatalln("Cannot get record from file. Aborting...")
		}
		page.records = append(page.records, temp_tree_record)
		i += 1
	}
	file.Close()
	return page
}
