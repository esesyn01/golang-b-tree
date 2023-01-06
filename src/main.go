package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	exit := 0
	root_address = int32(NO_ROOT)
	create_bin_file(TREE_FILE_NAME)
	create_bin_file(RECORDS_FILE_NAME)
	for scanner.Scan() {
		line := scanner.Text()
		args := strings.Fields(line)
		option_s := []byte(args[0])
		option := option_s[0]
		switch option {
		case 'I':
			key, _ := strconv.Atoi(args[1])
			if root_address == NO_ROOT {
				address := get_offset_for_new_record()
				create_new_record(int32(key), int32(address))
				create_tree(int32(key), address)
			} else {
				root := read_page_from_file(root_address)
				leaf, _, res := find_key(&root, int32(key))
				address := get_offset_for_new_record()
				if res == true {
					fmt.Printf("Record with %d key already exists\n", key)
				} else {
					create_new_record(int32(key), int32(address))
					insert_key(leaf, int32(key), int32(address), int32(NO_CHILD), int32(NO_CHILD))
				}
			}

		case 'F':
			key, _ := strconv.Atoi(args[1])
			root := read_page_from_file(root_address)
			fmt.Println(root_address)
			_, record_address, res := find_key(&root, int32(key))
			if res == true {
				found_record := read_record_from_file(record_address)
				log.Printf("Found record with key %d - ", key)
				fmt.Println(found_record)
			} else {
				log.Printf("Record with key %d not found\n", key)
			}
		case 'T':
			fmt.Println("Display tree")
		case 'R':
			fmt.Println("Display records")
		case 'D':
			fmt.Println("Delete node")
		case 'E':
			exit = 1
		default:
			fmt.Println("Bad option")
		}
		if exit == 1 {
			break
		}
	}
}
