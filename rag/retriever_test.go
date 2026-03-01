package rag

import (
	"testing"
)

// TestSimpleRetrieverAdd 测试添加文档
func TestSimpleRetrieverAdd(t *testing.T) {
	retriever := NewSimpleRetriever()
	doc := Document{
		ID:      "doc1",
		Content: "Hello world",
	}

	err := retriever.Add(doc)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestSimpleRetrieverAddEmptyID 测试空ID
func TestSimpleRetrieverAddEmptyID(t *testing.T) {
	retriever := NewSimpleRetriever()
	doc := Document{
		ID:      "",
		Content: "Hello world",
	}

	err := retriever.Add(doc)
	if err == nil {
		t.Error("Expected error for empty ID")
	}
}

// TestSimpleRetrieverSearch 测试搜索
func TestSimpleRetrieverSearch(t *testing.T) {
	retriever := NewSimpleRetriever()
	retriever.Add(Document{ID: "doc1", Content: "Go programming language"})
	retriever.Add(Document{ID: "doc2", Content: "Python programming language"})
	retriever.Add(Document{ID: "doc3", Content: "Java programming language"})

	results, err := retriever.Search("Go programming", 2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}

// TestSimpleRetrieverDelete 测试删除
func TestSimpleRetrieverDelete(t *testing.T) {
	retriever := NewSimpleRetriever()
	retriever.Add(Document{ID: "doc1", Content: "Hello world"})

	err := retriever.Delete("doc1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = retriever.Delete("doc1")
	if err == nil {
		t.Error("Expected error for deleting non-existent document")
	}
}

// TestSimpleRetrieverClear 测试清空
func TestSimpleRetrieverClear(t *testing.T) {
	retriever := NewSimpleRetriever()
	retriever.Add(Document{ID: "doc1", Content: "Hello world"})

	err := retriever.Clear()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	results, _ := retriever.Search("Hello", 5)
	if len(results) != 0 {
		t.Error("Expected no results after clear")
	}
}

// TestVectorRetrieverAdd 测试向量检索器添加
func TestVectorRetrieverAdd(t *testing.T) {
	retriever := NewVectorRetriever()
	doc := Document{
		ID:      "doc1",
		Content: "Machine learning algorithms",
	}

	err := retriever.Add(doc)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestVectorRetrieverSearch 测试向量搜索
func TestVectorRetrieverSearch(t *testing.T) {
	retriever := NewVectorRetriever()
	retriever.Add(Document{ID: "doc1", Content: "Machine learning algorithms"})
	retriever.Add(Document{ID: "doc2", Content: "Deep learning neural networks"})

	results, err := retriever.Search("learning", 2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}

// TestDocumentStore 测试文档存储
func TestDocumentStore(t *testing.T) {
	retriever := NewSimpleRetriever()
	store := NewDocumentStore(retriever)

	err := store.AddDocument(Document{ID: "doc1", Content: "Test document"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	results, err := store.Search("Test", 5)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}

// TestRetrieverInterface 测试接口实现
func TestRetrieverInterface(t *testing.T) {
	var _ Retriever = NewSimpleRetriever()
	var _ Retriever = NewVectorRetriever()
}

// BenchmarkSimpleRetrieverSearch 基准测试搜索
func BenchmarkSimpleRetrieverSearch(b *testing.B) {
	retriever := NewSimpleRetriever()
	for i := 0; i < 100; i++ {
		retriever.Add(Document{
			ID:      string(rune(i)),
			Content: "Test document content for benchmarking",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		retriever.Search("Test document", 5)
	}
}
