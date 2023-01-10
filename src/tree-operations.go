package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

func find_key(node *tree_page, key int32) (*tree_page, int32, bool) {
	if node == nil {
		return nil, -2, false
	}
	i := 0
	for int32(i) < node.header.records_number {
		if node.records[i].key == key {
			return node, node.records[i].record_offset, true
		}
		i += 1
	}
	res := is_leaf(node)
	if res == true {
		return node, -2, false
	}
	if key < node.records[0].key {
		child := read_page_from_file(node.first_child_offset)
		return find_key(&child, key)
	}
	if key > node.records[node.header.records_number-1].key {
		child := read_page_from_file(node.records[node.header.records_number-1].child_page_offset)
		return find_key(&child, key)
	}
	i = 0
	for int32(i) < node.header.records_number-1 {
		if key > node.records[i].key && key < node.records[i+1].key {
			child := read_page_from_file(node.records[i].child_page_offset)
			return find_key(&child, key)
		}
		i += 1
	}

	log.Printf("Key %d not found!\n", key)
	return node, -2, false
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

func insert_key(child *tree_page, key int32, address, child1, child2 int32) {
	var new_record tree_record
	new_record.child_page_offset = child2
	new_record.key = key
	new_record.record_offset = address
	if child.header.records_number < MAX_KEYS {
		insert_key_into_node(child, new_record, child1)
		write_page_to_file(*child)
		return
	}
	res := is_root(child)
	if res == true {
		split_root(child, &new_record, child1)
		return
	}
	parent := read_page_from_file(child.header.parent_offset)
	result := compensate(&parent, child, &new_record, child1)
	if result == true {
		//insert_key_into_leaf(child, key, address)
		return
	}
	new_key, new_record_address, new_page_adress := split(&parent, child, new_record, child1)
	insert_key(&parent, new_key, new_record_address, child.header.own_offset, new_page_adress)
	return
}

func split_root(old_root *tree_page, new_record *tree_record, child1 int32) {
	insert_key_into_node(old_root, *new_record, child1)
	var new_root tree_page
	var header tree_page_header
	header.records_number = 1
	header.own_offset = get_offset_for_new_node()
	new_root.first_child_offset = old_root.header.own_offset
	header.parent_offset = NO_PARENT
	new_root.header = header
	new_sibling := alloc_new_page(new_root)
	new_sibling.header.own_offset += TREE_PAGE_SIZE
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
	update_parent_offsets(&new_root, NO_PARENT)
	return
}

func compensate(parent, child *tree_page, new_record *tree_record, child1 int32) bool {
	fmt.Println("Siema")
	if new_record.key < parent.records[0].key {
		sibling := read_page_from_file(parent.records[0].child_page_offset)
		if sibling.header.records_number == MAX_KEYS {
			return false
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
		sibling.first_child_offset = values[pivot-1].child_page_offset
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
		return true
	}

	if new_record.key > parent.records[parent.header.records_number-1].key {
		var sibling tree_page
		if parent.header.records_number == int32(1) {
			sibling = read_page_from_file(parent.first_child_offset)
		} else {
			sibling = read_page_from_file(parent.records[parent.header.records_number-2].child_page_offset)
		}
		if sibling.header.records_number == MAX_KEYS {
			return false
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
		return true
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
					return false
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
				return true
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
				right.first_child_offset = values[pivot-1].child_page_offset
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
				return true
			}
		}
		i += 1
	}
	return false
}

func split(parent, child *tree_page, new_record tree_record, child1 int32) (int32, int32, int32) {
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
	update_parent_offsets(parent, parent.header.parent_offset)
	return values[pivot].key, values[pivot].record_offset, new_page.header.own_offset
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

func print_records(node tree_page) {
	res := is_leaf(&node)
	if res == true {
		i := 0
		for int32(i) < node.header.records_number {
			read_record := read_record_from_file(node.records[i].record_offset)
			fmt.Println(read_record)
			i += 1
		}
		return
	} else {
		child := read_page_from_file(node.first_child_offset)
		print_records(child)
		i := 0
		for int32(i) < node.header.records_number {
			read_record := read_record_from_file(node.records[i].record_offset)
			fmt.Println(read_record)
			child = read_page_from_file(node.records[i].child_page_offset)
			print_records(child)
			i += 1
		}
		return
	}
}

func print_tree(node tree_page, depth int, levels []string) []string {
	res := is_leaf(&node)
	if res == true {
		levels = print_node(node, depth, levels)
		return levels
	} else {
		levels = print_node(node, depth, levels)
		i := 0
		child := read_page_from_file(node.first_child_offset)
		levels = print_tree(child, depth+1, levels)
		for int32(i) < node.header.records_number {
			child = read_page_from_file(node.records[i].child_page_offset)
			levels = print_tree(child, depth+1, levels)
			i += 1
		}
		return levels
	}
}

func print_node(node tree_page, depth int, levels []string) []string {
	diff := tree_height - depth - 1
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
	for i < tree_height {
		levels[i] += "\n"
		fmt.Print(levels[i])
		i += 1
	}
	file.Close()
	return
}

func delete_tree(root *tree_page) {

}

func delete_key(node *tree_page, key int32) {
	res := is_leaf(node)
	if res == true {
		underflow := delete_key_from_leaf(node, key)
		if underflow == true {
			further_delete(node)
			return
		} else {
			return
		}
	} else {
		index := get_key_index_in_page(node, key)
		var leaf *tree_page
		var swap_key int32
		var underflow bool
		var swap_address int32
		if int32(index) == node.header.records_number-1 {
			if index == 0 {
				child := read_page_from_file(node.first_child_offset)
				leaf, underflow, swap_key, swap_address = del_max_key(&child)
			} else {
				child := read_page_from_file(node.records[index-1].child_page_offset)
				leaf, underflow, swap_key, swap_address = del_max_key(&child)
			}
		} else {
			child := read_page_from_file(node.records[index].child_page_offset)
			leaf, underflow, swap_key, swap_address = del_min_key(&child)
		}
		node.records[index].key = swap_key
		node.records[index].record_offset = swap_address
		if underflow == true {
			further_delete(leaf)
			return
		} else {
			return
		}
	}
}

func further_delete(node *tree_page) {
	res2 := is_root(node)
	if res2 == true && node.header.records_number == 0 {
		delete_tree(node)
		return
	} else if res2 == true && tree_height == 1 {
		return
	} else if res2 == false {
		parent := read_page_from_file(node.header.parent_offset)
		comp_possible := compensate_delete(node, &parent)
		if comp_possible == true {
			return
		} else {
			merge(node, &parent)
			underflow_parent := is_underflow(&parent)
			if underflow_parent == true {
				further_delete(&parent)
				return
			}
		}
	} else {
		merge_root(node)
		return
	}
}

func delete_key_from_leaf(node *tree_page, key int32) bool {
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
			node.records = remove_from_list(node.records, i)
			node.header.records_number -= 1
			return is_underflow(node)
		}
	}
	return false
}

func compensate_delete(child, parent *tree_page) bool {
	return false
}

func merge_root(old_root *tree_page) {

}

func merge(child, parent *tree_page) {

}
