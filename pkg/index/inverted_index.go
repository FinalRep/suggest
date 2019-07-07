package index

import (
	"fmt"
	"io"

	"github.com/alldroll/suggest/pkg/compression"
)

type (
	// DocumentID is a unique indentificator of a indexed document
	DocumentID = uint32
	// Term represents an independent search element in a search document
	Term = string
	// Position (posting) is a list item of PostingList
	Position = DocumentID
	// Index is a low level data structure for storing a map of posting lists
	Index = map[Term][]Position
)

// InvertedIndex is an index data structure that contains list of
// references to documents for each term
type InvertedIndex interface {
	// Get returns corresponding posting list for given term
	Get(term Term) (PostingListIterator, error)
	// Has checks is there is given term in inverted index
	Has(term Term) bool
}

// invertedIndexStructure is a first part of inverted index. It is stored in memory,
// with pointers to each posting list, which is stored on disk
type invertedIndexStructure map[Term]struct {
	// size - byte length of position list for given term
	size uint32
	// position - is file position, where positing list is stored
	position uint32
	// length - posting list length
	length uint32
}

// NewInvertedIndex returns new instance of InvertedIndex that is stored on disc
func NewInvertedIndex(
	reader Input,
	decoder compression.Decoder,
	table invertedIndexStructure,
) InvertedIndex {
	return &invertedIndex{
		reader:  reader,
		decoder: decoder,
		table:   table,
	}
}

// invertedIndex implements InvertedIndex interface
type invertedIndex struct {
	reader  Input
	decoder compression.Decoder
	table   invertedIndexStructure
}

// Get returns corresponding posting list for given term
func (i *invertedIndex) Get(term Term) (PostingListIterator, error) {
	s, ok := i.table[term]

	if !ok {
		return nil, nil
	}

	postingList := make([]Position, s.length)
	reader := io.NewSectionReader(i.reader, int64(s.position), int64(s.size))
	n, err := i.decoder.Decode(reader, postingList)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode posting list: %v", err)
	}

	return NewPostingListIterator(postingList[:n]), nil
}

// Has checks is there is given term in inverted index
func (i *invertedIndex) Has(term Term) bool {
	_, ok := i.table[term]
	return ok
}
