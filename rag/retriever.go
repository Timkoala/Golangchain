// Package rag 提供检索增强生成系统
package rag

import (
	"errors"
	"math"
	"sort"
	"strings"
	"sync"
)

// Document 定义文档
type Document struct {
	ID       string
	Content  string
	Metadata map[string]interface{}
}

// SearchResult 定义搜索结果
type SearchResult struct {
	Document Document
	Score    float64
}

// Retriever 定义检索器接口
type Retriever interface {
	Add(doc Document) error
	Search(query string, topK int) ([]SearchResult, error)
	Delete(docID string) error
	Clear() error
}

// SimpleRetriever 定义简单检索器
type SimpleRetriever struct {
	documents map[string]Document
	mu        sync.RWMutex
}

// NewSimpleRetriever 创建新的简单检索器
func NewSimpleRetriever() *SimpleRetriever {
	return &SimpleRetriever{
		documents: make(map[string]Document),
	}
}

// Add 添加文档
func (sr *SimpleRetriever) Add(doc Document) error {
	if doc.ID == "" {
		return errors.New("document ID cannot be empty")
	}
	if doc.Content == "" {
		return errors.New("document content cannot be empty")
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.documents[doc.ID] = doc
	return nil
}

// Search 搜索文档
func (sr *SimpleRetriever) Search(query string, topK int) ([]SearchResult, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}
	if topK <= 0 {
		topK = 5
	}

	sr.mu.RLock()
	defer sr.mu.RUnlock()

	results := make([]SearchResult, 0)

	for _, doc := range sr.documents {
		score := sr.calculateSimilarity(query, doc.Content)
		if score > 0 {
			results = append(results, SearchResult{
				Document: doc,
				Score:    score,
			})
		}
	}

	// 按分数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 返回前 topK 个结果
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// Delete 删除文档
func (sr *SimpleRetriever) Delete(docID string) error {
	if docID == "" {
		return errors.New("document ID cannot be empty")
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	if _, exists := sr.documents[docID]; !exists {
		return errors.New("document not found")
	}

	delete(sr.documents, docID)
	return nil
}

// Clear 清空所有文档
func (sr *SimpleRetriever) Clear() error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.documents = make(map[string]Document)
	return nil
}

// calculateSimilarity 计算相似度（基于词频）
func (sr *SimpleRetriever) calculateSimilarity(query, content string) float64 {
	queryWords := strings.Fields(strings.ToLower(query))
	contentWords := strings.Fields(strings.ToLower(content))

	if len(queryWords) == 0 || len(contentWords) == 0 {
		return 0
	}

	matchCount := 0
	for _, qWord := range queryWords {
		for _, cWord := range contentWords {
			if qWord == cWord {
				matchCount++
				break
			}
		}
	}

	// 计算 Jaccard 相似度
	return float64(matchCount) / float64(len(queryWords))
}

// VectorRetriever 定义向量检索器
type VectorRetriever struct {
	documents map[string]Document
	vectors   map[string][]float64
	mu        sync.RWMutex
}

// NewVectorRetriever 创建新的向量检索器
func NewVectorRetriever() *VectorRetriever {
	return &VectorRetriever{
		documents: make(map[string]Document),
		vectors:   make(map[string][]float64),
	}
}

// Add 添加文档
func (vr *VectorRetriever) Add(doc Document) error {
	if doc.ID == "" {
		return errors.New("document ID cannot be empty")
	}
	if doc.Content == "" {
		return errors.New("document content cannot be empty")
	}

	vr.mu.Lock()
	defer vr.mu.Unlock()

	vr.documents[doc.ID] = doc
	// 简单的向量化：基于词频
	vr.vectors[doc.ID] = vr.vectorize(doc.Content)
	return nil
}

// Search 搜索文档
func (vr *VectorRetriever) Search(query string, topK int) ([]SearchResult, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}
	if topK <= 0 {
		topK = 5
	}

	vr.mu.RLock()
	defer vr.mu.RUnlock()

	queryVector := vr.vectorize(query)
	results := make([]SearchResult, 0)

	for docID, docVector := range vr.vectors {
		score := vr.cosineSimilarity(queryVector, docVector)
		if score > 0 {
			results = append(results, SearchResult{
				Document: vr.documents[docID],
				Score:    score,
			})
		}
	}

	// 按分数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 返回前 topK 个结果
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// Delete 删除文档
func (vr *VectorRetriever) Delete(docID string) error {
	if docID == "" {
		return errors.New("document ID cannot be empty")
	}

	vr.mu.Lock()
	defer vr.mu.Unlock()

	if _, exists := vr.documents[docID]; !exists {
		return errors.New("document not found")
	}

	delete(vr.documents, docID)
	delete(vr.vectors, docID)
	return nil
}

// Clear 清空所有文档
func (vr *VectorRetriever) Clear() error {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	vr.documents = make(map[string]Document)
	vr.vectors = make(map[string][]float64)
	return nil
}

// vectorize 向量化文本
func (vr *VectorRetriever) vectorize(text string) []float64 {
	words := strings.Fields(strings.ToLower(text))
	wordFreq := make(map[string]float64)

	for _, word := range words {
		wordFreq[word]++
	}

	// 转换为向量（简化版本）
	vector := make([]float64, 0)
	for _, freq := range wordFreq {
		vector = append(vector, freq)
	}

	return vector
}

// cosineSimilarity 计算余弦相似度
func (vr *VectorRetriever) cosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) == 0 || len(vec2) == 0 {
		return 0
	}

	minLen := len(vec1)
	if len(vec2) < minLen {
		minLen = len(vec2)
	}

	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0

	for i := 0; i < minLen; i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// DocumentStore 定义文档存储
type DocumentStore struct {
	retriever Retriever
	mu        sync.RWMutex
}

// NewDocumentStore 创建新的文档存储
func NewDocumentStore(retriever Retriever) *DocumentStore {
	return &DocumentStore{
		retriever: retriever,
	}
}

// AddDocument 添加文档
func (ds *DocumentStore) AddDocument(doc Document) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	return ds.retriever.Add(doc)
}

// Search 搜索文档
func (ds *DocumentStore) Search(query string, topK int) ([]SearchResult, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.retriever.Search(query, topK)
}

// DeleteDocument 删除文档
func (ds *DocumentStore) DeleteDocument(docID string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	return ds.retriever.Delete(docID)
}

// Clear 清空存储
func (ds *DocumentStore) Clear() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	return ds.retriever.Clear()
}
