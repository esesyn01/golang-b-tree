package main

const (
	TREE_DEGREE             = 2
	HEADER_SIZE             = 12
	SINGLE_RECORD_SIZE      = 16
	SINGLE_TREE_RECORD_SIZE = 12
	INT_SIZE                = 4
	MAX_KEYS                = 2 * TREE_DEGREE
	MAX_CHILDREN            = 2*TREE_DEGREE + 1
	TREE_PAGE_SIZE          = HEADER_SIZE + MAX_KEYS*SINGLE_TREE_RECORD_SIZE + INT_SIZE
	TREE_FILE_NAME          = "bin/tree.bin"
	RECORDS_FILE_NAME       = "bin/records.bin"
	NO_CHILD                = -3
	NO_PARENT               = -4
	NO_ROOT                 = -5
	CHARS_FOR_NODE          = 3*MAX_KEYS + 1*MAX_CHILDREN
	SPACE_PADDING           = 1
	TREE_GRAPH_FILE         = "bin/graph.txt"
	COMMANDS_FILE           = "bin/commands.txt"
	DELETED                 = -6
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
	own_offset     int32
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

var root_address int32
var tree_height int
var free_list_records []int32
var free_list_pages []int32
