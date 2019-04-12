package index

import (
	"encoding/gob"
	"fmt"
	"runtime"

	"github.com/alldroll/suggest/pkg/compression"
)

// Reader is an entity, providing access to a search index
type Reader struct {
	directory Directory
	config    WriterConfig
	decoder   compression.Decoder
}

// NewIndexReader returns a new instance of a search index reader
func NewIndexReader(
	directory Directory,
	config WriterConfig,
	decoder compression.Decoder,
) *Reader {
	return &Reader{
		directory: directory,
		config:    config,
		decoder:   decoder,
	}
}

// Read reads a inverted index indices from the given directory
func (ir *Reader) Read() (InvertedIndexIndices, error) {
	// read header
	header, err := ir.readHeader()

	if err != nil {
		return nil, err
	}

	documentReader, err := ir.directory.OpenInput(ir.config.DocumentListFileName)

	if err != nil {
		return nil, fmt.Errorf("Failed to open document list: %v", err)
	}

	index, err := ir.createInvertedIndexIndices(header, documentReader)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve inverted index: %v", err)
	}

	runtime.SetFinalizer(index, func(d interface{}) {
		closeIfRequired(documentReader)
	})

	return index, nil
}

// readHeader reads an index header from the given directory
func (ir *Reader) readHeader() (*header, error) {
	headerReader, err := ir.directory.OpenInput(ir.config.HeaderFileName)

	if err != nil {
		return nil, fmt.Errorf("Failed to open header: %v", err)
	}

	header := &header{}
	decoder := gob.NewDecoder(headerReader)

	if err = decoder.Decode(header); err != nil {
		return nil, fmt.Errorf("Failed to retrieve header: %v", err)
	}

	if err = closeIfRequired(headerReader); err != nil {
		return nil, fmt.Errorf("Failed to close header file: %v", err)
	}

	return header, nil
}

// createInvertedIndexIndices creates new instance of InvertedIndexIndices from the given header
func (ir *Reader) createInvertedIndexIndices(header *header, documentReader Input) (InvertedIndexIndices, error) {
	// create inverted index structure slice
	indices := make([]InvertedIndex, int(header.Indices))
	invertedIndexStructureIndices := make([]invertedIndexStructure, len(indices))

	// here we create list of invertedIndexStructure
	for _, description := range header.Terms {
		if description.PostingListSize == 0 {
			invertedIndexStructureIndices[description.Indice] = nil
			continue
		}

		if invertedIndexStructureIndices[description.Indice] == nil {
			invertedIndexStructureIndices[description.Indice] = make(invertedIndexStructure)
		}

		invertedIndexStructureIndices[description.Indice][description.Term] = struct {
			size     uint32
			position uint32
		}{size: description.PostingListSize, position: description.PostingListPosition}
	}

	// create NewInvertedIndex for given indice
	for i, invertedIndexStructure := range invertedIndexStructureIndices {
		if invertedIndexStructure == nil {
			indices[i] = nil
		} else {
			indices[i] = NewInvertedIndex(documentReader, ir.decoder, invertedIndexStructure)
		}
	}

	return NewInvertedIndexIndices(indices), nil
}