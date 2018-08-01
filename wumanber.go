package wumanber

import (
	"errors"
	"os"
	"bytes"
	"encoding/binary"
	"log"
)

func HashCode(str string) uint32 {
	var hash uint32 = 0
	for i := range str {
		hash = uint32(str[i]) + (hash << 6) + (hash << 16) - hash
	}
	return hash & 0x7FFFFFFF
}

type PrefixIdPair struct {
	Hash uint32
	Index int32
}

type PrefixTable []PrefixIdPair


type WuManber struct {
	// minum length of patterns
	Min int32
	// SHIFT table
	ShiftTable []int32
	// a combination of HASH and PREFIX table
	HashTable []PrefixTable
	// patterns
	Patterns []string
	// size of SHIFT and HASH table
	TableSize int32
	// size of block
	Block int32
}

func (w *WuManber) Init(patterns []string) error {
	// init block
	w.Block = 3

	patternSize := int32(len(patterns))
	if patternSize == 0 {
		return errors.New("WuManber init failed because no pattern specified")
	}

	w.Patterns = make([]string, patternSize)

	w.Min = int32(len(patterns[0]))
	var lenPattern int32 = 0
	for i, p := range patterns {
		w.Patterns[i] = p
		//fmt.Println(p)
		lenPattern = int32(len(p))
		if lenPattern < w.Min {
			w.Min = lenPattern
		}
	}

	if w.Block > w.Min {
		log.Println("Warning: Block is larger than minum pattern length, reset mBlock to minmum, but it will seriously affect the effiency.")
		w.Block = w.Min
	}

	primes := []int32 {1003, 10007, 100003, 1000003, 10000019, 100000007}
	threshold := 10 * w.Min
	for _, p := range primes {
		if p > patternSize && p / patternSize > threshold {
			w.TableSize = p
			break
		}
	}

	if w.TableSize == 0 {
		log.Println("Warning: amount of pattern is very large, will cost a great amount of memory.")
		w.TableSize = primes[5]
	}
	//fmt.Println(w.TableSize)


	w.HashTable = make([]PrefixTable, w.TableSize)
	for i := 0; i < int(w.TableSize); i++ {
		w.HashTable[i] = make(PrefixTable, 0)
	}

	defaultValue := w.Min - w.Block + 1
	w.ShiftTable = make([]int32, w.TableSize)
	for i := range w.ShiftTable {
		w.ShiftTable[i] = defaultValue
	}

	for id := range patterns {
		for index := w.Min; index >= w.Block; index-- {
			start := index - w.Block
			//fmt.Println(patterns[id][start:start + w.Block])
			hashCode := HashCode(patterns[id][start:start + w.Block]) % uint32(w.TableSize)
			//fmt.Println(hashCode)
			if w.ShiftTable[hashCode] > (w.Min - index) {
				w.ShiftTable[hashCode] = w.Min - index
			}
			if index == w.Min {
				prefixHash := HashCode(patterns[id][0:w.Block])
				//prefixHash
				w.HashTable[hashCode] = append(w.HashTable[hashCode], PrefixIdPair{prefixHash, int32(id)})
			}
		}
	}
	return nil
}

func (w *WuManber) Search(text string) int {
	// hit count
	var hits int = 0
	var index int32 = w.Min - 1; // start off by matching end of largest common pattern

	var blockMaxIndex int32 = w.Block - 1
	var windowMaxIndex int32 = w.Min - 1

	textLength := int32(len(text))
	for index < textLength {
		blockHash := HashCode(text[index - blockMaxIndex: index - blockMaxIndex + w.Block])
		blockHash = blockHash % uint32(w.TableSize)
		shift := w.ShiftTable[blockHash]
		if shift > 0 {
			index += shift
		} else {
			prefixHash := HashCode(text[index - windowMaxIndex:index-windowMaxIndex+w.Block])
			var p = &(w.HashTable[blockHash])
			for _, pp := range *p {
				if prefixHash == pp.Hash {
					// since prefindex matches, compare target substring with pattern
					// we know first two characters already match
					lenPattern := len(w.Patterns[pp.Index])
					var i = index - windowMaxIndex
					var j = 0
					for ; i < textLength && j < lenPattern; {
						if w.Patterns[pp.Index][j] != text[i] {
							break
						}
						i++
						j++
					}
					if j == lenPattern {
						hits++
						//log.Println(w.Patterns[pp.Index])
					}
				}
			} // end for
			index++
		} // end else
	} // end for
	return hits
}

func (w *WuManber) Serialize(path string) error {
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		return errors.New(err.Error())
	}
	var binBuf bytes.Buffer
	binary.Write(&binBuf, binary.BigEndian, w.Min)
	binary.Write(&binBuf, binary.BigEndian, w.TableSize)
	binary.Write(&binBuf, binary.BigEndian, w.Block)

	// write SHIFT table to buffer
	//for i := range w.ShiftTable {
	//	binary.Write(&binBuf, binary.BigEndian, w.ShiftTable[i])
	//}
	binary.Write(&binBuf, binary.BigEndian, w.ShiftTable)

	// write Hash table to buffer
	for i := range w.HashTable {
		//fmt.Println(len(w.HashTable[i]))
		binary.Write(&binBuf, binary.BigEndian, int32(len(w.HashTable[i])))
		for j := range w.HashTable[i] {
			binary.Write(&binBuf, binary.BigEndian, w.HashTable[i][j])
		}
	}

	binary.Write(&binBuf, binary.BigEndian, int32(len(w.Patterns)))
	for i := range w.Patterns {
		binary.Write(&binBuf, binary.BigEndian, int32(len(w.Patterns[i])))
		binBuf.Write([]byte(w.Patterns[i]))
	}
	file.Write(binBuf.Bytes())
	log.Println("Serialize WuManber model successfully.")
	return nil
}


func (w *WuManber) Deserialize(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.New(err.Error())
	}
	data := readNextBytes(file, 4)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &w.Min)

	//fmt.Println(w.Min)

	data = readNextBytes(file, 4)
	buffer = bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &w.TableSize)

	//fmt.Println(w.TableSize)

	data = readNextBytes(file, 4)
	buffer = bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &w.Block)

	//fmt.Println(w.Block)

	w.ShiftTable = make([]int32, w.TableSize)
	data = readNextBytes(file, 4 * int(w.TableSize))
	buffer = bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &w.ShiftTable)

	log.Println("successfully deserialize SHIFT table")

	w.HashTable = make([]PrefixTable, w.TableSize)
	var sizeOfPrefixIdPair int = 8
	for i := 0; i < int(w.TableSize); i++ {
		var l int32
		data = readNextBytes(file, 4)
		buffer = bytes.NewBuffer(data)
		err = binary.Read(buffer, binary.BigEndian, &l)

		w.HashTable[i] = make([]PrefixIdPair, l)
		data = readNextBytes(file, sizeOfPrefixIdPair * int(l))
		buffer = bytes.NewBuffer(data)
		err = binary.Read(buffer, binary.BigEndian, &w.HashTable[i])
	}

	log.Println("successfully deserialize Hash table")

	var patternSize int32 = 0

	data = readNextBytes(file, 4)
	buffer = bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &patternSize)

	w.Patterns = make([]string, patternSize)
	for i := 0; i < int(patternSize); i++ {
		var l int32 = 0
		data = readNextBytes(file, 4)
		buffer = bytes.NewBuffer(data)
		err = binary.Read(buffer, binary.BigEndian, &l)
		data = readNextBytes(file, int(l))
		buffer = bytes.NewBuffer(data)
		//err = binary.Read(buffer, binary.BigEndian, &w.Patterns[i])
		w.Patterns[i] = string(buffer.Bytes())
	}
	log.Println("successfully deserialize patterns")
	return nil
}

func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}
