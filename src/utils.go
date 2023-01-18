package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func insert_into_list(list []tree_record, index int, value tree_record) []tree_record {
	if len(list) == index {
		return append(list, value)
	}
	list = append(list[:index+1], list[index:]...)
	list[index] = value
	return list
}

func remove_from_list(slice []tree_record, s int) []tree_record {

	if len(slice)-1 == s {
		return slice[:len(slice)-1]
	}
	if s == 0 {
		return slice[1:]
	}
	return append(slice[:s], slice[s+1:]...)
}

func is_leaf(page *tree_page) bool {
	if page.first_child_offset == int32(NO_CHILD) {
		return true
	}
	return false
}

func is_root(page *tree_page) bool {
	if page.header.parent_offset != NO_PARENT {
		return false
	}
	return true
}

func is_underflow(page *tree_page) bool {
	if page.header.records_number < TREE_DEGREE {
		return true
	}
	return false
}

func alloc_new_page(parent tree_page) tree_page {
	var new_page tree_page
	var header tree_page_header
	header.own_offset = get_offset_for_new_node()
	header.records_number = TREE_DEGREE
	header.parent_offset = parent.header.own_offset
	new_page.header = header
	new_page.first_child_offset = NO_CHILD
	return new_page
}

func update_parent_offsets(node *tree_page, offset int32) int {
	operations := 0
	if node.header.parent_offset != offset {
		node.header.parent_offset = offset
		operations += 1
	}
	res := is_leaf(node)
	if res == false {
		child := read_page_from_file(node.first_child_offset)
		operations += 1
		operations += update_parent_offsets(&child, node.header.own_offset)
		i := 0
		for int32(i) < node.header.records_number {
			child = read_page_from_file(node.records[i].child_page_offset)
			operations += update_parent_offsets(&child, node.header.own_offset)
			operations += 1
			i += 1
		}
	}
	write_page_to_file(*node)
	return operations
}

func del_max_key(node *tree_page, path []tree_page) (*tree_page, bool, int32, int32, []tree_page) {
	res := is_leaf(node)
	path = append(path, *node)
	if res == true {
		key := node.records[node.header.records_number-1].key
		address := node.records[node.header.records_number-1].record_offset
		underflow := delete_key_from_leaf(node, key)
		return node, underflow, key, address, path
	}
	child := read_page_from_file(node.records[node.header.records_number-1].child_page_offset)
	return del_max_key(&child, path)
}

func del_min_key(node *tree_page, path []tree_page) (*tree_page, bool, int32, int32, []tree_page) {
	res := is_leaf(node)
	path = append(path, *node)
	if res == true {
		key := node.records[0].key
		address := node.records[0].record_offset
		underflow := delete_key_from_leaf(node, key)
		return node, underflow, key, address, path
	}
	child := read_page_from_file(node.first_child_offset)
	return del_min_key(&child, path)
}

func get_key_index_in_page(node *tree_page, key int32) int {
	i := 0
	for int32(i) < node.header.records_number {
		if key == node.records[i].key {
			return i
		}
		i += 1
	}
	return -1
}

func delete_key_from_leaf(node *tree_page, key int32) bool {
	i := 0
	for int32(i) < node.header.records_number {
		if key == node.records[i].key {
			node.records = remove_from_list(node.records, i)
			node.header.records_number -= 1
			return is_underflow(node)
		}
		i += 1
	}
	return false
}

func insert_key_into_node(node *tree_page, new_tree_record tree_record, offset int32) {
	if new_tree_record.key < node.records[0].key {
		node.records = insert_into_list(node.records, 0, new_tree_record)
		node.header.records_number += 1
		node.first_child_offset = offset
		return
	}
	if new_tree_record.key > node.records[node.header.records_number-1].key {
		node.records = insert_into_list(node.records, int(node.header.records_number), new_tree_record)
		node.header.records_number += 1
		return
	}
	i := 0
	for int32(i) < node.header.records_number-1 {
		if new_tree_record.key > node.records[i].key && new_tree_record.key < node.records[i+1].key {
			node.records = insert_into_list(node.records, i+1, new_tree_record)
			node.header.records_number += 1
			return
		}
		i += 1
	}

	log.Fatalln("Cannot insert record into leaf. Aborting...")
}

func generate_test_data() {
	file := get_file(COMMANDS_FILE)
	i := 0
	for i < 200 {
		rand.Seed(time.Now().UTC().UnixNano())
		key := (rand.Int31() % 299) + 1
		fmt.Fprintf(file, "I %d\n", key)
		i += 1
	}
	for i < 12000 {
		rand.Seed(time.Now().UTC().UnixNano())
		key := (rand.Int31() % 999) + 1
		opt := rand.Int31() % 2
		if opt == 0 {
			fmt.Fprintf(file, "D %d\n", key)
		} else {
			fmt.Fprintf(file, "I %d\n", key)
		}
		i += 1
	}
	file.Close()
}

func test_tree(node tree_page, elems []int32) []int32 {
	res := is_leaf(&node)
	operations := 0
	if res == true {
		i := 0
		for int32(i) < node.header.records_number {
			read_record := read_record_from_file(node.records[i].record_offset)
			elems = append(elems, read_record.key)
			if read_record.key == DELETED {
				log.Fatalln("Unused key in node!")
			}
			i += 1
		}
		return elems
	} else {
		child := read_page_from_file(node.first_child_offset)
		operations += 1
		elems = test_tree(child, elems)
		i := 0
		for int32(i) < node.header.records_number {
			read_record := read_record_from_file(node.records[i].record_offset)
			elems = append(elems, read_record.key)
			if read_record.key == DELETED {
				log.Fatalln("Unused key in node!")
			}
			child = read_page_from_file(node.records[i].child_page_offset)
			operations += 1
			elems = test_tree(child, elems)
			i += 1
		}
		return elems
	}
}

func analzye_elems(elems []int32) {
	i := 1
	for i < len(elems) {
		if elems[i] <= elems[i-1] {
			log.Fatalf("Keys %d and %d are not in order\n", elems[i-1], elems[i])
		}
		i += 1
	}
	return
}

func print_subtree(node tree_page) {
	levels := []string{}
	i := 0
	for i < subtree_height {
		new_string := ""
		levels = append(levels, new_string)
		i += 1
	}
	levels, _ = print_tree(node, 0, levels)
	print_result_tree(levels)
}

func calc_used_space(node tree_page) (int, int) {
	records := int(node.header.records_number)
	nodes := 1
	res := is_leaf(&node)
	if res == true {
		return records, nodes
	}
	i := 0
	child := read_page_from_file(node.first_child_offset)
	new_records, new_nodes := calc_used_space(child)
	records += new_records
	nodes += new_nodes
	for int32(i) < node.header.records_number {
		child = read_page_from_file(node.records[i].child_page_offset)
		new_records, new_nodes = calc_used_space(child)
		records += new_records
		nodes += new_nodes
		i += 1
	}
	return records, nodes
}
