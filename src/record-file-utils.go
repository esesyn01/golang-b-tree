package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	"time"
)

func write_record_to_file(record_to_store record, offset int32) {
	file := get_file(RECORDS_FILE_NAME)
	off := int64(offset)
	_, err := file.Seek(off, 0)
	if err != nil {
		file.Close()
		log.Fatalln("Cannot achieve desired offset. Aborting...")
	}
	err = binary.Write(file, binary.LittleEndian, record_to_store)
	if err != nil {
		file.Close()
		fmt.Println(err)
		log.Fatalln("Cannot write record to file. Aborting...")
	}
	file.Close()
}

func read_record_from_file(offset int32) record {
	file := get_file(RECORDS_FILE_NAME)
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
	buffer1 := bytes.NewBuffer(buffer[0:INT_SIZE])
	buffer2 := bytes.NewBuffer(buffer[INT_SIZE : 2*INT_SIZE])
	buffer3 := bytes.NewBuffer(buffer[2*INT_SIZE : 3*INT_SIZE])
	buffer4 := bytes.NewBuffer(buffer[3*INT_SIZE : 4*INT_SIZE])

	err = binary.Read(buffer1, binary.LittleEndian, &read_record.mass)
	if err != nil {
		log.Fatalln("Cannot read record field")
	}
	err = binary.Read(buffer2, binary.LittleEndian, &read_record.heat)
	if err != nil {
		log.Fatalln("Cannot read record field")
	}
	err = binary.Read(buffer3, binary.LittleEndian, &read_record.temp_delta)
	if err != nil {
		log.Fatalln("Cannot read record field")
	}
	err = binary.Read(buffer4, binary.LittleEndian, &read_record.key)
	if err != nil {
		log.Fatalln("Cannot read record field")
	}
	file.Close()
	return read_record
}

func create_new_record(key, address int32) {
	rand.Seed(time.Now().UTC().UnixNano())
	var new_record record
	new_record.mass = (rand.Int31() % 999) + 1
	new_record.heat = (rand.Int31() % 999) + 1
	new_record.temp_delta = (rand.Int31() % 999) + 1
	new_record.key = key
	write_record_to_file(new_record, address)
}
func delete_record_from_file(node tree_page, key int32) {
	i := 0
	for int32(i) < node.header.records_number {
		if key == node.records[i].key {
			del_record := read_record_from_file(node.records[i].record_offset)
			del_record.key = DELETED
			del_record.temp_delta = DELETED
			del_record.mass = DELETED
			del_record.heat = DELETED
			free_list_records = append(free_list_records, node.records[i].record_offset)
			write_record_to_file(del_record, node.records[i].record_offset)
			return
		}
		i += 1
	}
}

func get_offset_for_new_record() int32 {
	if len(free_list_records) > 0 {
		addr := free_list_records[len(free_list_records)-1]
		free_list_records = free_list_records[:len(free_list_records)-1]
		return addr
	} else {
		file := get_file(RECORDS_FILE_NAME)
		fileinfo, _ := file.Stat()
		size := fileinfo.Size()
		file.Close()
		return int32(size)
	}
}

func update_record(address int32) {
	updated_record := read_record_from_file(address)
	fmt.Println("Pass value for mass")
	fmt.Scanf("%d", &updated_record.mass)
	fmt.Println("Pass value for heat")
	fmt.Scanf("%d", &updated_record.heat)
	fmt.Println("Pass value for delta temperature")
	fmt.Scanf("%d", &updated_record.temp_delta)
	write_record_to_file(updated_record, address)

}
