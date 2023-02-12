package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

func find_key(node *tree_page, key int32, path []tree_page) (*tree_page, int32, bool, []tree_page) {
	if node == nil {
		return nil, -2, false, []tree_page{}
	}
	path = append(path, *node)
	i := 0
	for int32(i) < node.header.records_number {
		if node.records[i].key == key {
			return node, node.records[i].record_offset, true, path
		}
		i += 1
	}
	res := is_leaf(node)
	if res == true {
		return node, -2, false, path
	}
	if key < node.records[0].key {
		child := read_page_from_file(node.first_child_offset)
		return find_key(&child, key, path)
	}
	if key > node.records[node.header.records_number-1].key {
		child := read_page_from_file(node.records[node.header.records_number-1].child_page_offset)
		return find_key(&child, key, path)
	}
	i = 0
	for int32(i) < node.header.records_number-1 {
		if key > node.records[i].key && key < node.records[i+1].key {
			child := read_page_from_file(node.records[i].child_page_offset)
			return find_key(&child, key, path)
		}
		i += 1
	}

	log.Printf("Key %d not found!\n", key)
	return node, -2, false, path
}

func create_tree(key, address int32) {
	var root tree_page
	var header tree_page_header
	header.own_offset = 0
	header.records_number = 1
	header.parent_offset = NO_PARENT
	root.header = header
	root.first_child_offset = NO_CHILD
	var new_record tree_record
	new_record.key = key
	new_record.record_offset = address
	new_record.child_page_offset = NO_CHILD
	root.records = append(root.records, new_record)
	root_address = root.header.own_offset
	tree_height = 1
	write_page_to_file(root)
	return

}

func insert_key(child *tree_page, key int32, address, child1, child2 int32, path []tree_page) int {
	var new_record tree_record
	new_record.child_page_offset = child2
	new_record.key = key
	new_record.record_offset = address
	operations := 0
	if child.header.records_number < MAX_KEYS {
		insert_key_into_node(child, new_record, child1)
		write_page_to_file(*child)
		return 1
	}
	//fmt.Println("only 'leaf'")
	//print_subtree(*child)
	res := is_root(child)
	if res == true {
		operations += split_root(child, &new_record, child1)
		return operations
	}
	//parent := read_page_from_file(child.header.parent_offset)
	parent := path[len(path)-2]
	subtree_height += 1
	//log.Println("Initial parent")
	//print_subtree(parent)
	result, new_ops := compensate(&parent, child, &new_record, child1)
	//log.Println("After companesate")
	//print_subtree(parent)
	operations += new_ops
	if result == true {
		//insert_key_into_leaf(child, key, address)
		return operations
	}
	new_key, new_record_address, new_page_adress, new_ops2 := split(&parent, child, new_record, child1)
	//log.Println("After split")
	//print_subtree(parent)
	operations += new_ops2
	path = path[:len(path)-1]
	operations += insert_key(&parent, new_key, new_record_address, child.header.own_offset, new_page_adress, path)
	return operations
}

func split_root(old_root *tree_page, new_record *tree_record, child1 int32) int {
	insert_key_into_node(old_root, *new_record, child1)
	var new_root tree_page
	var header tree_page_header
	header.records_number = 1
	header.own_offset = get_offset_for_new_node()
	new_root.first_child_offset = old_root.header.own_offset
	header.parent_offset = NO_PARENT
	new_root.header = header
	new_sibling := alloc_new_page(new_root)
	if new_sibling.header.own_offset == new_root.header.own_offset {
		new_sibling.header.own_offset += TREE_PAGE_SIZE
	}
	values := old_root.records
	old_root.header.parent_offset = new_root.header.own_offset
	old_root.header.records_number = TREE_DEGREE
	old_root.records = []tree_record{}
	pivot := int(len(values) / 2)
	i := 0
	for i < pivot {
		old_root.records = append(old_root.records, values[i])
		i += 1
	}
	new_root.records = append(new_root.records, values[pivot])
	new_root.records[0].child_page_offset = new_sibling.header.own_offset
	new_sibling.first_child_offset = values[pivot].child_page_offset
	i = pivot + 1
	for i < len(values) {
		new_sibling.records = append(new_sibling.records, values[i])
		i += 1
	}
	write_page_to_file(*old_root)
	write_page_to_file(new_root)
	write_page_to_file(new_sibling)
	root_address = new_root.header.own_offset
	tree_height += 1
	operations := update_parent_offsets(&new_root, NO_PARENT)
	return operations + 3
}

func compensate(parent, child *tree_page, new_record *tree_record, child1 int32) (bool, int) {
	if new_record.key < parent.records[0].key {
		sibling := read_page_from_file(parent.records[0].child_page_offset)
		if sibling.header.records_number == MAX_KEYS {
			return false, 1
		}
		insert_key_into_node(child, *new_record, child1)
		values := child.records
		values = append(values, parent.records[0])
		parent_pos := len(values) - 1
		i := 0
		for int32(i) < sibling.header.records_number {
			values = append(values, sibling.records[i])
			i += 1
		}
		pivot := int(len(values) / 2)
		//diff := pivot - parent_pos
		parent.records[0].key = values[pivot].key
		parent.records[0].record_offset = values[pivot].record_offset
		values[parent_pos].child_page_offset = sibling.first_child_offset
		sibling.first_child_offset = values[pivot].child_page_offset
		i = 0
		child.records = []tree_record{}
		child.header.records_number = int32(pivot)
		for i < pivot {
			child.records = append(child.records, values[i])
			i += 1
		}
		i = pivot + 1
		sibling.header.records_number = int32(len(values) - pivot - 1)
		sibling.records = []tree_record{}
		for i < len(values) {
			sibling.records = append(sibling.records, values[i])
			i += 1
		}
		write_page_to_file(sibling)
		write_page_to_file(*child)
		write_page_to_file(*parent)
		return true, 4
	}

	if new_record.key > parent.records[parent.header.records_number-1].key {
		var sibling tree_page
		if parent.header.records_number == int32(1) {
			sibling = read_page_from_file(parent.first_child_offset)
		} else {
			sibling = read_page_from_file(parent.records[parent.header.records_number-2].child_page_offset)
		}
		if sibling.header.records_number == MAX_KEYS {
			return false, 1
		}
		insert_key_into_node(child, *new_record, child1)
		values := sibling.records
		values = append(values, parent.records[parent.header.records_number-1])
		parent_pos := len(values) - 1
		i := 0
		for int32(i) < child.header.records_number {
			values = append(values, child.records[i])
			i += 1
		}
		pivot := int(len(values) / 2)
		//diff := pivot - parent_pos
		parent.records[parent.header.records_number-1].key = values[pivot].key
		parent.records[parent.header.records_number-1].record_offset = values[pivot].record_offset
		values[parent_pos].child_page_offset = child.first_child_offset
		child.first_child_offset = values[pivot].child_page_offset
		i = 0
		sibling.records = []tree_record{}
		sibling.header.records_number = int32(pivot)
		for i < pivot {
			sibling.records = append(sibling.records, values[i])
			i += 1
		}
		i = pivot + 1
		child.header.records_number = int32(len(values) - pivot - 1)
		child.records = []tree_record{}
		for i < len(values) {
			child.records = append(child.records, values[i])
			i += 1
		}
		write_page_to_file(sibling)
		write_page_to_file(*child)
		write_page_to_file(*parent)
		return true, 4
	}

	i := 0
	for int32(i) < parent.header.records_number-1 {
		if new_record.key > parent.records[i].key && new_record.key < parent.records[i+1].key {
			k := i
			var left tree_page
			if i == 0 {
				left = read_page_from_file(parent.first_child_offset)
			} else {
				left = read_page_from_file(parent.records[i-1].child_page_offset)
			}
			right := read_page_from_file(parent.records[i+1].child_page_offset)
			if right.header.records_number == MAX_KEYS {
				if left.header.records_number == MAX_KEYS {
					return false, 2
				}
				insert_key_into_node(child, *new_record, child1)
				values := left.records
				values = append(values, parent.records[k])
				parent_pos := len(values) - 1
				i2 := 0
				for int32(i2) < child.header.records_number {
					values = append(values, child.records[i2])
					i2 += 1
				}
				pivot := int(len(values) / 2)
				//diff := pivot - parent_pos
				parent.records[k].key = values[pivot].key
				parent.records[k].record_offset = values[pivot].record_offset
				values[parent_pos].child_page_offset = child.first_child_offset
				child.first_child_offset = values[pivot].child_page_offset
				i2 = 0
				left.records = []tree_record{}
				left.header.records_number = int32(pivot)
				for i2 < pivot {
					left.records = append(left.records, values[i2])
					i2 += 1
				}
				i2 = pivot + 1
				child.header.records_number = int32(len(values) - pivot - 1)
				child.records = []tree_record{}
				for i2 < len(values) {
					child.records = append(child.records, values[i2])
					i2 += 1
				}
				write_page_to_file(left)
				write_page_to_file(*child)
				write_page_to_file(*parent)
				return true, 5
			} else {
				insert_key_into_node(child, *new_record, child1)
				values := child.records
				values = append(values, parent.records[k+1])
				parent_pos := len(values) - 1
				i2 := 0
				for int32(i2) < right.header.records_number {
					values = append(values, right.records[i2])
					i2 += 1
				}
				pivot := int(len(values) / 2)
				//diff := pivot - parent_pos
				parent.records[k+1].key = values[pivot].key
				parent.records[k+1].record_offset = values[pivot].record_offset
				values[parent_pos].child_page_offset = right.first_child_offset
				right.first_child_offset = values[pivot].child_page_offset
				i2 = 0
				child.records = []tree_record{}
				child.header.records_number = int32(pivot)
				for i2 < pivot {
					child.records = append(child.records, values[i2])
					i2 += 1
				}
				i2 = pivot + 1
				right.header.records_number = int32(len(values) - pivot - 1)
				right.records = []tree_record{}
				for i2 < len(values) {
					right.records = append(right.records, values[i2])
					i2 += 1
				}
				write_page_to_file(right)
				write_page_to_file(*child)
				write_page_to_file(*parent)
				return true, 5
			}
		}
		i += 1
	}
	return false, 0
}

func split(parent, child *tree_page, new_record tree_record, child1 int32) (int32, int32, int32, int) {
	new_page := alloc_new_page(*parent)
	insert_key_into_node(child, new_record, child1)
	pivot := int(child.header.records_number / 2)
	values := child.records
	child.records = []tree_record{}
	new_page.header.records_number = TREE_DEGREE
	child.header.records_number = TREE_DEGREE
	i := 0
	for i < pivot {
		child.records = append(child.records, values[i])
		i += 1
	}
	i = pivot + 1
	new_page.first_child_offset = values[pivot].child_page_offset
	for i < len(values) {
		new_page.records = append(new_page.records, values[i])
		i += 1
	}
	write_page_to_file(*parent)
	write_page_to_file(*child)
	write_page_to_file(new_page)
	operations := update_parent_offsets(parent, parent.header.parent_offset)
	return values[pivot].key, values[pivot].record_offset, new_page.header.own_offset, operations + 3
}

func print_records(node tree_page) int {
	res := is_leaf(&node)
	operations := 0
	if res == true {
		i := 0
		for int32(i) < node.header.records_number {
			read_record := read_record_from_file(node.records[i].record_offset)
			fmt.Println(read_record)
			i += 1
		}
		return operations
	} else {
		child := read_page_from_file(node.first_child_offset)
		operations += 1
		operations += print_records(child)
		i := 0
		for int32(i) < node.header.records_number {
			read_record := read_record_from_file(node.records[i].record_offset)
			fmt.Println(read_record)
			child = read_page_from_file(node.records[i].child_page_offset)
			operations += 1
			operations += print_records(child)
			i += 1
		}
		return operations
	}
}

func print_tree(node tree_page, depth int, levels []string) ([]string, int) {
	res := is_leaf(&node)
	operations := 0
	if res == true {
		levels = print_node(node, depth, levels)
		return levels, operations
	} else {
		levels = print_node(node, depth, levels)
		i := 0
		child := read_page_from_file(node.first_child_offset)
		operations += 1
		var new_ops int
		levels, new_ops = print_tree(child, depth+1, levels)
		operations += new_ops
		for int32(i) < node.header.records_number {
			child = read_page_from_file(node.records[i].child_page_offset)
			operations += 1
			levels, new_ops = print_tree(child, depth+1, levels)
			operations += new_ops
			i += 1
		}
		return levels, operations
	}
}

func print_node(node tree_page, depth int, levels []string) []string {
	diff := subtree_height - depth - 1
	var needed_spaces int
	if diff == 0 {
		needed_spaces = 1
	} else {
		needed_spaces = int(int(math.Pow(MAX_CHILDREN, float64(diff-1))) * SPACE_PADDING * CHARS_FOR_NODE / 2)
	}
	node_string := strings.Repeat(" ", needed_spaces)
	node_string += "[ ."
	node_string += strings.Repeat(" ", needed_spaces)
	i := 0
	for int32(i) < node.header.records_number {
		node_string += strconv.Itoa(int(node.records[i].key))
		node_string += strings.Repeat(" ", needed_spaces)
		if int32(i+1) == node.header.records_number {
			node_string += ". ]"
		} else {
			node_string += "."
			node_string += strings.Repeat(" ", needed_spaces)
		}
		i += 1
	}
	levels[depth] += node_string
	return levels
}

func print_result_tree(levels []string) {
	file := get_file(TREE_GRAPH_FILE)
	i := 0
	for i < subtree_height {
		levels[i] += "\n"
		fmt.Print(levels[i])
		i += 1
	}
	file.Close()
	return
}

func fprint_result_tree(levels []string) {
	file := get_file(TREE_GRAPH_FILE)
	i := 0
	for i < subtree_height {
		levels[i] += "\n"
		fmt.Fprint(file, levels[i])
		i += 1
	}
	file.Close()
	return
}

func delete_tree() {
	tree_height = 0
	root_address = int32(NO_ROOT)
	remove_file(TREE_FILE_NAME)
	remove_file(RECORDS_FILE_NAME)
	create_bin_file(TREE_FILE_NAME)
	create_bin_file(RECORDS_FILE_NAME)
	free_list_pages = []int32{}
	free_list_records = []int32{}
	return
}

func delete_key(node *tree_page, key int32, path []tree_page) int {
	res := is_leaf(node)
	operations := 0
	if res == true {
		underflow := delete_key_from_leaf(node, key)
		delete_record_from_file(*node, key)
		write_page_to_file(*node)
		operations += 1
		if underflow == true {
			operations += further_delete(node, path)
			return operations
		} else {
			return operations
		}
	} else {
		index := get_key_index_in_page(node, key)
		var leaf *tree_page
		var swap_key int32
		var underflow bool
		var swap_address int32
		var add_path []tree_page
		if int32(index) == node.header.records_number-1 {
			if index == 0 {
				child := read_page_from_file(node.first_child_offset)
				leaf, underflow, swap_key, swap_address, add_path = del_max_key(&child, add_path)
			} else {
				child := read_page_from_file(node.records[index-1].child_page_offset)
				leaf, underflow, swap_key, swap_address, add_path = del_max_key(&child, add_path)
			}
		} else {
			child := read_page_from_file(node.records[index].child_page_offset)
			leaf, underflow, swap_key, swap_address, add_path = del_min_key(&child, add_path)
		}
		operations += len(add_path)
		i := 0
		for i < len(add_path) {
			path = append(path, add_path[i])
			i += 1
		}
		write_page_to_file(*leaf)
		delete_record_from_file(*node, key)
		node.records[index].key = swap_key
		node.records[index].record_offset = swap_address
		write_page_to_file(*node)
		operations += 2
		if underflow == true {
			operations += further_delete(leaf, path)
			return operations
		} else {
			return operations
		}
	}
}

func further_delete(node *tree_page, path []tree_page) int {
	res2 := is_root(node)
	operations := 0
	if res2 == true && node.header.records_number == 0 {
		delete_tree()
		return 1
	} else if res2 == true && tree_height == 1 {
		return operations
	} else if res2 == false {
		parent := path[len(path)-2]
		//parent := read_page_from_file(node.header.parent_offset)
		comp_possible, sibling, new_ops := compensate_delete(node, &parent)
		operations += new_ops
		if comp_possible == true {
			return operations
		} else {
			res3 := is_root(&parent)
			if res3 == true && parent.header.records_number == 1 {
				operations += merge_root(&parent, node)
				return operations
			}
			operations += merge(node, &parent, sibling)
			underflow_parent := is_underflow(&parent)
			if underflow_parent == true {
				path = path[:len(path)-1]
				operations += further_delete(&parent, path)
				return operations
			}
		}
	}
	return 0
}

func compensate_delete(child, parent *tree_page) (bool, *tree_page, int) {
	if child.records[0].key < parent.records[0].key {
		sibling := read_page_from_file(parent.records[0].child_page_offset)
		if sibling.header.records_number == TREE_DEGREE {
			return false, &sibling, 1
		}
		values := child.records
		values = append(values, parent.records[0])
		parent_pos := len(values) - 1
		i := 0
		for int32(i) < sibling.header.records_number {
			values = append(values, sibling.records[i])
			i += 1
		}
		pivot := int(len(values) / 2)
		//diff := pivot - parent_pos
		parent.records[0].key = values[pivot].key
		parent.records[0].record_offset = values[pivot].record_offset
		values[parent_pos].child_page_offset = sibling.first_child_offset
		sibling.first_child_offset = values[pivot].child_page_offset
		i = 0
		child.records = []tree_record{}
		child.header.records_number = int32(pivot)
		for i < pivot {
			child.records = append(child.records, values[i])
			i += 1
		}
		i = pivot + 1
		sibling.header.records_number = int32(len(values) - pivot - 1)
		sibling.records = []tree_record{}
		for i < len(values) {
			sibling.records = append(sibling.records, values[i])
			i += 1
		}
		write_page_to_file(sibling)
		write_page_to_file(*child)
		write_page_to_file(*parent)
		return true, &sibling, 4
	}

	if child.records[0].key > parent.records[parent.header.records_number-1].key {
		var sibling tree_page
		if parent.header.records_number == int32(1) {
			sibling = read_page_from_file(parent.first_child_offset)
		} else {
			sibling = read_page_from_file(parent.records[parent.header.records_number-2].child_page_offset)
		}
		if sibling.header.records_number == TREE_DEGREE {
			return false, &sibling, 1
		}
		values := sibling.records
		values = append(values, parent.records[parent.header.records_number-1])
		parent_pos := len(values) - 1
		i := 0
		for int32(i) < child.header.records_number {
			values = append(values, child.records[i])
			i += 1
		}
		pivot := int(len(values) / 2)
		//diff := pivot - parent_pos
		parent.records[parent.header.records_number-1].key = values[pivot].key
		parent.records[parent.header.records_number-1].record_offset = values[pivot].record_offset
		values[parent_pos].child_page_offset = child.first_child_offset
		child.first_child_offset = values[pivot].child_page_offset
		i = 0
		sibling.records = []tree_record{}
		sibling.header.records_number = int32(pivot)
		for i < pivot {
			sibling.records = append(sibling.records, values[i])
			i += 1
		}
		i = pivot + 1
		child.header.records_number = int32(len(values) - pivot - 1)
		child.records = []tree_record{}
		for i < len(values) {
			child.records = append(child.records, values[i])
			i += 1
		}
		write_page_to_file(sibling)
		write_page_to_file(*child)
		write_page_to_file(*parent)
		return true, &sibling, 4
	}

	i := 0
	var right tree_page
	for int32(i) < parent.header.records_number-1 {
		if child.records[0].key > parent.records[i].key && child.records[0].key < parent.records[i+1].key {
			k := i
			var left tree_page
			if i == 0 {
				left = read_page_from_file(parent.first_child_offset)
			} else {
				left = read_page_from_file(parent.records[i-1].child_page_offset)
			}
			right = read_page_from_file(parent.records[i+1].child_page_offset)
			if right.header.records_number == TREE_DEGREE {
				if left.header.records_number == TREE_DEGREE {
					return false, &left, 2
				}
				values := left.records
				values = append(values, parent.records[k])
				parent_pos := len(values) - 1
				i2 := 0
				for int32(i2) < child.header.records_number {
					values = append(values, child.records[i2])
					i2 += 1
				}
				pivot := int(len(values) / 2)
				//diff := pivot - parent_pos
				parent.records[k].key = values[pivot].key
				parent.records[k].record_offset = values[pivot].record_offset
				values[parent_pos].child_page_offset = child.first_child_offset
				child.first_child_offset = values[pivot].child_page_offset
				i2 = 0
				left.records = []tree_record{}
				left.header.records_number = int32(pivot)
				for i2 < pivot {
					left.records = append(left.records, values[i2])
					i2 += 1
				}
				i2 = pivot + 1
				child.header.records_number = int32(len(values) - pivot - 1)
				child.records = []tree_record{}
				for i2 < len(values) {
					child.records = append(child.records, values[i2])
					i2 += 1
				}
				write_page_to_file(left)
				write_page_to_file(*child)
				write_page_to_file(*parent)
				return true, &left, 5
			} else {
				values := child.records
				values = append(values, parent.records[k+1])
				parent_pos := len(values) - 1
				i2 := 0
				for int32(i2) < right.header.records_number {
					values = append(values, right.records[i2])
					i2 += 1
				}
				pivot := int(len(values) / 2)
				//diff := pivot - parent_pos
				parent.records[k+1].key = values[pivot].key
				parent.records[k+1].record_offset = values[pivot].record_offset
				values[parent_pos].child_page_offset = right.first_child_offset
				right.first_child_offset = values[pivot].child_page_offset
				i2 = 0
				child.records = []tree_record{}
				child.header.records_number = int32(pivot)
				for i2 < pivot {
					child.records = append(child.records, values[i2])
					i2 += 1
				}
				i2 = pivot + 1
				right.header.records_number = int32(len(values) - pivot - 1)
				right.records = []tree_record{}
				for i2 < len(values) {
					right.records = append(right.records, values[i2])
					i2 += 1
				}
				write_page_to_file(right)
				write_page_to_file(*child)
				write_page_to_file(*parent)
				return true, &left, 5
			}
		}
		i += 1
	}
	return false, &right, 0
}

func merge_root(old_root, child *tree_page) int {
	var sibling tree_page
	var values []tree_record
	var parent_index int
	operations := 0
	if old_root.records[0].key > child.records[0].key {
		sibling = read_page_from_file(old_root.records[0].child_page_offset)
		operations += 1
		values = child.records
		values = append(values, old_root.records[0])
		parent_index = len(values) - 1
		values[parent_index].child_page_offset = sibling.first_child_offset
		i := 0
		for int32(i) < sibling.header.records_number {
			values = append(values, sibling.records[i])
			i += 1
		}
	} else {
		sibling = read_page_from_file(old_root.first_child_offset)
		operations += 1
		old_child_first_offset := child.first_child_offset
		child.first_child_offset = sibling.first_child_offset
		values = sibling.records
		values = append(values, old_root.records[0])
		parent_index = len(values) - 1
		values[parent_index].child_page_offset = old_child_first_offset
		i := 0
		for int32(i) < child.header.records_number {
			values = append(values, child.records[i])
			i += 1
		}
	}

	child.records = values
	child.header.records_number = MAX_KEYS
	child.header.parent_offset = NO_PARENT
	root_address = child.header.own_offset
	tree_height -= 1

	sibling.records = []tree_record{}
	sibling.header.parent_offset = DELETED
	sibling.first_child_offset = DELETED
	sibling.header.records_number = DELETED
	old_root.records = []tree_record{}
	old_root.header.parent_offset = DELETED
	old_root.first_child_offset = DELETED
	old_root.header.records_number = DELETED
	operations += update_parent_offsets(child, child.header.parent_offset)
	write_page_to_file(*old_root)
	write_page_to_file(*child)
	write_page_to_file(sibling)
	free_list_pages = append(free_list_pages, sibling.header.own_offset)
	free_list_pages = append(free_list_pages, old_root.header.own_offset)
	return operations + 3
}

func merge(child, parent, sibling *tree_page) int {
	var values []tree_record
	var parent_index int
	var parent_idx_in_values int
	operations := 0
	i := 0
	if child.records[0].key < sibling.records[0].key {
		values = child.records
		for int32(i) < parent.header.records_number {
			if child.records[0].key < parent.records[i].key && sibling.records[0].key > parent.records[i].key {
				parent_index = i
				break
			}
			i += 1
		}
		values = append(values, parent.records[parent_index])
		parent_idx_in_values = len(values) - 1
		values[parent_idx_in_values].child_page_offset = sibling.first_child_offset
		i = 0
		for int32(i) < sibling.header.records_number {
			values = append(values, sibling.records[i])
			i += 1
		}

	} else {
		values = sibling.records
		old_child_first_offset := child.first_child_offset
		child.first_child_offset = sibling.first_child_offset
		for int32(i) < parent.header.records_number {
			if child.records[0].key > parent.records[i].key && sibling.records[0].key < parent.records[i].key {
				parent_index = i
				break
			}
			i += 1
		}
		values = append(values, parent.records[parent_index])
		parent_idx_in_values = len(values) - 1
		values[parent_idx_in_values].child_page_offset = old_child_first_offset
		i = 0
		for int32(i) < child.header.records_number {
			values = append(values, child.records[i])
			i += 1
		}
		if parent_index == 0 {
			parent.first_child_offset = child.header.own_offset
		} else {
			parent.records[parent_index-1].child_page_offset = child.header.own_offset
		}
	}

	parent.records = remove_from_list(parent.records, parent_index)
	parent.header.records_number -= 1
	sibling.records = []tree_record{}
	sibling.header.parent_offset = DELETED
	sibling.first_child_offset = DELETED
	sibling.header.records_number = DELETED
	child.records = values
	child.header.records_number = MAX_KEYS
	operations += update_parent_offsets(parent, parent.header.parent_offset)
	write_page_to_file(*parent)
	write_page_to_file(*child)
	write_page_to_file(*sibling)
	free_list_pages = append(free_list_pages, sibling.header.own_offset)
	return operations + 3
}
