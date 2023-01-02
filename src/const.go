package main

const (
	TREE_DEGREE             = 2
	HEADER_SIZE             = 8
	SINGLE_RECORD_SIZE      = 16
	SINGLE_TREE_RECORD_SIZE = 12
	TREE_PAGE_SIZE          = HEADER_SIZE + 2*TREE_DEGREE*SINGLE_TREE_RECORD_SIZE + 4
)

type record struct {
	mass       int32
	heat       int32
	temp_delta int32
	key        int32
}

type tree_page_header struct {
	parent_offset  int32
	records_number int32
}

type tree_record struct {
	key               int32
	record_offset     int32
	child_page_offset int32
}

type tree_page struct {
	header             tree_page_header
	first_child_offset int32
	records            []tree_record
}
