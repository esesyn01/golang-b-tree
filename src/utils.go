package main

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

func update_parent_offsets(node *tree_page, offset int32) {
	node.header.parent_offset = offset
	res := is_leaf(node)
	if res == false {
		child := read_page_from_file(node.first_child_offset)
		update_parent_offsets(&child, node.header.own_offset)
		i := 0
		for int32(i) < node.header.records_number {
			child = read_page_from_file(node.records[i].child_page_offset)
			update_parent_offsets(&child, node.header.own_offset)
			i += 1
		}
	}
	write_page_to_file(*node)
	return
}

func del_max_key(node *tree_page) (*tree_page, bool, int32, int32) {
	res := is_leaf(node)
	if res == true {
		key := node.records[node.header.records_number-1].key
		address := node.records[node.header.records_number-1].record_offset
		underflow := delete_key_from_leaf(node, key)
		return node, underflow, key, address
	}
	child := read_page_from_file(node.records[node.header.records_number-1].child_page_offset)
	return del_max_key(&child)
}

func del_min_key(node *tree_page) (*tree_page, bool, int32, int32) {
	res := is_leaf(node)
	if res == true {
		key := node.records[0].key
		address := node.records[0].record_offset
		underflow := delete_key_from_leaf(node, key)
		return node, underflow, key, address
	}
	child := read_page_from_file(node.first_child_offset)
	return del_min_key(&child)
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
