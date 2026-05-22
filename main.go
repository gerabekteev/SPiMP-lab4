package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"
)

// cleanWord переводит слово в нижний регистр и очищает от начальных/конечных знаков препинания.
func cleanWord(w string) string {
	w = strings.ToLower(w)
	return strings.TrimFunc(w, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

// processChunk обрабатывает виртуальный фрагмент файла по заданным смещениям.
func processChunk(filePath string, startOffset, endOffset int64, wg *sync.WaitGroup, resultsChan chan<- map[string]struct{}) {
	defer wg.Done()

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Ошибка при открытии файла: %v\n", err)
		resultsChan <- nil
		return
	}
	defer file.Close()

	_, err = file.Seek(startOffset, io.SeekStart)
	if err != nil {
		fmt.Printf("Ошибка перемещения в файле: %v\n", err)
		resultsChan <- nil
		return
	}

	const bufSize = 1024 * 1024 // Буфер 1 МБ
	buf := make([]byte, bufSize)
	localMap := make(map[string]struct{})
	wordBuf := make([]byte, 0, 256)

	var currentOffset int64 = startOffset
	skippingFirstWord := startOffset > 0

	for {
		n, err := file.Read(buf)
		if n == 0 {
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("Ошибка чтения файла: %v\n", err)
				break
			}
		}

		for i := 0; i < n; i++ {
			b := buf[i]
			offset := currentOffset + int64(i)

			isDelim := b == ' ' || b == '\t' || b == '\r' || b == '\n'

			if skippingFirstWord {
				if isDelim {
					skippingFirstWord = false
				}
				continue
			}

			if !isDelim {
				wordBuf = append(wordBuf, b)
			} else {
				if len(wordBuf) > 0 {
					word := string(wordBuf)
					word = cleanWord(word)
					if word != "" {
						localMap[word] = struct{}{}
					}
					wordBuf = wordBuf[:0]
				}

				if offset >= endOffset {
					resultsChan <- localMap
					return
				}
			}
		}

		currentOffset += int64(n)
		if err == io.EOF {
			break
		}
	}

	// Обработка последнего слова перед EOF, если файл закончился без разделителя
	if len(wordBuf) > 0 && !skippingFirstWord {
		word := string(wordBuf)
		word = cleanWord(word)
		if word != "" {
			localMap[word] = struct{}{}
		}
	}

	resultsChan <- localMap
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: wordcounter <путь_к_файлу>")
		return
	}
	filePath := os.Args[1]

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("Ошибка получения информации о файле: %v\n", err)
		return
	}
	fileSize := fileInfo.Size()

	numCPU := runtime.NumCPU()
	fmt.Printf("Размер файла: %.2f ГБ (%d байт)\n", float64(fileSize)/(1024*1024*1024), fileSize)
	fmt.Printf("Количество доступных ядер CPU: %d\n", numCPU)

	startTime := time.Now()

	var wg sync.WaitGroup
	resultsChan := make(chan map[string]struct{}, numCPU)

	chunkSize := fileSize / int64(numCPU)

	for i := 0; i < numCPU; i++ {
		startOffset := int64(i) * chunkSize
		endOffset := int64(i+1) * chunkSize
		if i == numCPU-1 {
			endOffset = fileSize
		}

		wg.Add(1)
		go processChunk(filePath, startOffset, endOffset, &wg, resultsChan)
	}

	// Закрываем канал после завершения всех горутин
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Объединяем результаты в мастер-карту
	masterMap := make(map[string]struct{})
	for localMap := range resultsChan {
		if localMap != nil {
			for word := range localMap {
				masterMap[word] = struct{}{}
			}
		}
	}

	elapsed := time.Since(startTime)

	fmt.Printf("Обработка завершена!\n")
	fmt.Printf("Количество уникальных слов: %d\n", len(masterMap))
	fmt.Printf("Время выполнения: %v\n", elapsed)
}
