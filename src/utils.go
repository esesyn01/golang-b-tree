package main

func insert_into_list(list []tree_record, index int, value tree_record) []tree_record {
	if len(list) == index {
		return append(list, value)
	}
	list = append(list[:index+1], list[index:]...)
	list[index] = value
	return list
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
