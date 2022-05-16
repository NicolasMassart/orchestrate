package utils

import (
	"fmt"
)

const NotFoundErr = "not found"
const emptyListErr = "empty list"

type sortedListItem struct {
	value interface{}
	key   int64 // @TODO Support key as interface
	next  *sortedListItem
}

type SortedList struct {
	head *sortedListItem
}

func NewSortedList() *SortedList {
	return &SortedList{}
}

func (l *SortedList) Append(value interface{}, key int64) int {
	idx := 0
	curValue := l.head
	nextValue := &sortedListItem{
		key:   key,
		value: value,
	}

	for {
		if curValue == nil {
			break
		}
		// @TODO Support desc/asc order
		if curValue.key > key {
			break
		}

		curValue = curValue.next
		idx++
	}

	if curValue == nil {
		l.head = nextValue
	} else {
		curValue.next = nextValue
	}

	return idx
}

func (l *SortedList) Head() interface{} {
	if l.head == nil {
		return nil
	}

	return l.head.next
}

func (l *SortedList) GetByKey(key int64) (interface{}, error) {
	if l.head == nil {
		return nil, nil
	}

	curValue := l.head
	for {
		if curValue == nil {
			break
		}

		if curValue.key >= key {
			break
		}

		curValue = l.head.next
	}

	if curValue == nil || curValue.key != key {
		return nil, nil
	}

	return curValue, nil
}

func (l *SortedList) RemoveByKey(key int64) error {
	if l.head == nil {
		return fmt.Errorf(emptyListErr)
	}

	if l.head.key == key {
		l.head = l.head.next
		return nil
	}

	curValue := l.head
	for {
		if curValue == nil || curValue.next == nil {
			break
		}

		if curValue.next.key >= key {
			break
		}

		curValue = l.head.next
	}

	if curValue == nil || curValue.next == nil || curValue.next.key != key {
		return fmt.Errorf(NotFoundErr)
	}

	curValue.next = curValue.next.next
	return nil
}
