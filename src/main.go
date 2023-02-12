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
	create_bin_file(COMMANDS_FILE)
	commands := get_file(COMMANDS_FILE)
	scanner2 := bufio.NewScanner(commands)
	exit := 0
	root_address = int32(NO_ROOT)
	tree_height = 0
	total_disk_accesses := 0
	total_operations := 0
	create_bin_file(TREE_FILE_NAME)
	create_bin_file(RECORDS_FILE_NAME)
	create_bin_file(TREE_GRAPH_FILE)
	generate_test_data()
	for scanner2.Scan() {
		if exit == 1 {
			break
		}
		line := scanner2.Text()
		args := strings.Fields(line)
		option_s := []byte(args[0])
		option := option_s[0]
		var operations int
		switch option {
		case 'I':
			operations = 0
			key, _ := strconv.Atoi(args[1])
			if root_address == NO_ROOT {
				address := get_offset_for_new_record()
				create_new_record(int32(key), int32(address))
				create_tree(int32(key), address)
				operations += 1
			} else {
				subtree_height = 1
				root := read_page_from_file(root_address)
				leaf, _, res, path := find_key(&root, int32(key), []tree_page{})
				operations += len(path)
				address := get_offset_for_new_record()
				if res == true {
					fmt.Printf("Record with %d key already exists, performed %d operations\n", key, operations)
				} else {
					create_new_record(int32(key), int32(address))
					operations += insert_key(leaf, int32(key), int32(address), int32(NO_CHILD), int32(NO_CHILD), path)
					fmt.Printf("Record with key %d inserted in %d file operations\n", key, operations)
				}
			}
			operations = 0
			total_operations -= 1
		case 'S':
			key, _ := strconv.Atoi(args[1])
			if root_address != NO_ROOT {
				root := read_page_from_file(root_address)
				_, record_address, res, path := find_key(&root, int32(key), []tree_page{})
				operations = len(path)
				if res == true {
					found_record := read_record_from_file(record_address)
					log.Printf("Found record with key %d - ", key)
					fmt.Println(found_record)
				} else {
					log.Printf("Record with key %d not found\n", key)
				}
				fmt.Printf("Operation find %d key ended in %d file operations\n", key, len(path))
			} else {
				fmt.Println("There is no b-tree created")
			}
		case 'T':
			if root_address != NO_ROOT {
				subtree_height = tree_height
				levels := []string{}
				i := 0
				for i < tree_height {
					new_string := ""
					levels = append(levels, new_string)
					i += 1
				}
				root := read_page_from_file(root_address)
				levels, operations = print_tree(root, 0, levels)
				print_result_tree(levels)
				fmt.Printf("Displayed tree performing %d file operations\n", operations+1)
			} else {
				fmt.Println("There is no b-tree created")
			}
		case 'F':
			if root_address != NO_ROOT {
				subtree_height = tree_height
				levels := []string{}
				i := 0
				for i < tree_height {
					new_string := ""
					levels = append(levels, new_string)
					i += 1
				}
				root := read_page_from_file(root_address)
				levels, operations = print_tree(root, 0, levels)
				fprint_result_tree(levels)
				fmt.Printf("Displayed tree performing %d file operations\n", operations+1)
			} else {
				fmt.Println("There is no b-tree created")
			}
		case 'R':
			if root_address != NO_ROOT {
				root := read_page_from_file(root_address)
				operations = print_records(root)
				fmt.Printf("Displayed records performing %d file operations\n", operations+1)
			} else {
				fmt.Println("There is no b-tree created")
			}
		case 'U':
			if root_address != NO_ROOT {
				operations = 0
				key, _ := strconv.Atoi(args[1])
				root := read_page_from_file(root_address)
				node, _, res, path := find_key(&root, int32(key), []tree_page{})
				operations += len(path)
				if res == true {
					i := 0
					for int32(i) < node.header.records_number {
						if node.records[i].key == int32(key) {
							update_record(node.records[i].record_offset)
						}
						i += 1
					}
					fmt.Printf("Record with key %d updated in %d file operations\n", key, operations)
				} else {
					fmt.Printf("Key %d not found, made %d operations\n", key, operations)
				}
			} else {
				fmt.Println("There is no b-tree created")
			}

		case 'D':
			if root_address != NO_ROOT {
				operations = 0
				key, _ := strconv.Atoi(args[1])
				root := read_page_from_file(root_address)
				leaf, _, res, path := find_key(&root, int32(key), []tree_page{})
				operations += len(path)
				if res == true {
					operations += delete_key(leaf, int32(key), path)
					fmt.Printf("Deleted record with %d key in %d file operations\n", key, operations)
				} else {
					fmt.Printf("Key %d doesn't exist, performed %d operations!\n", key, operations)
				}
			} else {
				fmt.Println("There is no b-tree created")
			}

		case 'E':
			exit = 1
			total_operations -= 1
		default:
			fmt.Println("Bad option")
			total_operations -= 1
		}
		total_disk_accesses += operations
		total_operations += 1
		if exit == 1 {
			break
		}
	}

	for scanner.Scan() {
		if exit == 1 {
			break
		}
		line := scanner.Text()
		args := strings.Fields(line)
		option_s := []byte(args[0])
		option := option_s[0]
		var operations int
		switch option {
		case 'I':
			operations = 0
			key, _ := strconv.Atoi(args[1])
			if root_address == NO_ROOT {
				address := get_offset_for_new_record()
				create_new_record(int32(key), int32(address))
				create_tree(int32(key), address)
				operations += 1
			} else {
				subtree_height = 1
				root := read_page_from_file(root_address)
				leaf, _, res, path := find_key(&root, int32(key), []tree_page{})
				operations += len(path)
				address := get_offset_for_new_record()
				if res == true {
					fmt.Printf("Record with %d key already exists, performed %d operations\n", key, operations)
				} else {
					create_new_record(int32(key), int32(address))
					operations += insert_key(leaf, int32(key), int32(address), int32(NO_CHILD), int32(NO_CHILD), path)
					fmt.Printf("Record with key %d inserted in %d file operations\n", key, operations)
				}
			}

		case 'S':
			key, _ := strconv.Atoi(args[1])
			if root_address != NO_ROOT {
				root := read_page_from_file(root_address)
				_, record_address, res, path := find_key(&root, int32(key), []tree_page{})
				operations = len(path)
				if res == true {
					found_record := read_record_from_file(record_address)
					log.Printf("Found record with key %d - ", key)
					fmt.Println(found_record)
				} else {
					log.Printf("Record with key %d not found\n", key)
				}
				fmt.Printf("Operation find %d key ended in %d file operations\n", key, len(path))
			} else {
				fmt.Println("There is no b-tree created")
			}
		case 'T':
			if root_address != NO_ROOT {
				subtree_height = tree_height
				levels := []string{}
				i := 0
				for i < tree_height {
					new_string := ""
					levels = append(levels, new_string)
					i += 1
				}
				root := read_page_from_file(root_address)
				levels, operations = print_tree(root, 0, levels)
				print_result_tree(levels)
				fmt.Printf("Displayed tree performing %d file operations\n", operations+1)
			} else {
				fmt.Println("There is no b-tree created")
			}
		case 'F':
			if root_address != NO_ROOT {
				subtree_height = tree_height
				levels := []string{}
				i := 0
				for i < tree_height {
					new_string := ""
					levels = append(levels, new_string)
					i += 1
				}
				root := read_page_from_file(root_address)
				levels, operations = print_tree(root, 0, levels)
				fprint_result_tree(levels)
				fmt.Printf("Displayed tree performing %d file operations\n", operations+1)
			} else {
				fmt.Println("There is no b-tree created")
			}
		case 'R':
			if root_address != NO_ROOT {
				root := read_page_from_file(root_address)
				operations = print_records(root)
				fmt.Printf("Displayed records performing %d file operations\n", operations+1)
			} else {
				fmt.Println("There is no b-tree created")
			}
		case 'U':
			if root_address != NO_ROOT {
				operations = 0
				key, _ := strconv.Atoi(args[1])
				root := read_page_from_file(root_address)
				node, _, res, path := find_key(&root, int32(key), []tree_page{})
				operations += len(path)
				if res == true {
					i := 0
					for int32(i) < node.header.records_number {
						if node.records[i].key == int32(key) {
							update_record(node.records[i].record_offset)
						}
						i += 1
					}
					fmt.Printf("Record with key %d updated in %d file operations\n", key, operations)
				} else {
					fmt.Printf("Key %d not found, made %d operations\n", key, operations)
				}
			} else {
				fmt.Println("There is no b-tree created")
			}

		case 'D':
			if root_address != NO_ROOT {
				operations = 0
				key, _ := strconv.Atoi(args[1])
				root := read_page_from_file(root_address)
				leaf, _, res, path := find_key(&root, int32(key), []tree_page{})
				operations += len(path)
				if res == true {
					operations += delete_key(leaf, int32(key), path)
					fmt.Printf("Deleted record with %d key in %d file operations\n", key, operations)
				} else {
					fmt.Printf("Key %d doesn't exist, performed %d operations!\n", key, operations)
				}
			} else {
				fmt.Println("There is no b-tree created")
			}

		case 'E':
			exit = 1
			total_operations -= 1
		default:
			fmt.Println("Bad option")
			total_operations -= 1
		}
		total_disk_accesses += operations
		total_operations += 1
		if exit == 1 {
			break
		}
	}
	root := read_page_from_file(root_address)
	records, nodes := calc_used_space(root)
	percentage := (float64(records) / float64(nodes*MAX_KEYS)) * 100
	fmt.Printf("Overall, the tree has %.2f%% memory usage\n", percentage)
	avg_disk_operations := (float64(total_disk_accesses) / float64(total_operations))
	fmt.Printf("Overall, there were %.2f disk accesses on average operation\n", avg_disk_operations)
	fmt.Printf("There are %d records in tree at the end\n", records)
	commands.Close()
	remove_file(COMMANDS_FILE)
}
