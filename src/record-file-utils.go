package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

func write_record_to_file(filename string, record_to_store record, offset int) {
	file := get_file(filename)
	off := int64(offset)
	_, err := file.Seek(off, 0)
	if err != nil {
		file.Close()
		log.Fatalln("Cannot achieve desired offset. Aborting...")
	}
	err = binary.Write(file, binary.LittleEndian, record_to_store)
	if err != nil {
		file.Close()
		log.Fatalln("Cannot write record to file. Aborting...")
	}
	file.Close()
}

func read_record_from_file(filename string, offset int) record {
	file := get_file(filename)
	buffer := make([]byte, SINGLE_RECORD_SIZE)
	_, err := file.ReadAt(buffer, int64(offset))
	{
		if err != nil {
			if err != io.EOF {
				log.Fatalln(err)
			}
		}
	}
	var read_record record
	buffer1 := bytes.NewBuffer(buffer[offset : offset+INT_SIZE])
	buffer2 := bytes.NewBuffer(buffer[offset+INT_SIZE : offset+2*INT_SIZE])
	buffer3 := bytes.NewBuffer(buffer[offset+2*INT_SIZE : offset+3*INT_SIZE])
	buffer4 := bytes.NewBuffer(buffer[offset+3*INT_SIZE : offset+4*INT_SIZE])

	err = binary.Read(buffer1, binary.LittleEndian, read_record.mass)
	if err != nil {
		log.Fatalln("Cannot read record field")
	}
	err = binary.Read(buffer2, binary.LittleEndian, read_record.heat)
	if err != nil {
		log.Fatalln("Cannot read record field")
	}
	err = binary.Read(buffer3, binary.LittleEndian, read_record.temp_delta)
	if err != nil {
		log.Fatalln("Cannot read record field")
	}
	err = binary.Read(buffer4, binary.LittleEndian, read_record.key)
	if err != nil {
		log.Fatalln("Cannot read record field")
	}
	file.Close()
	return read_record
}
