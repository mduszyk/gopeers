package peers

import "math/big"

type bucketList struct {
	k int
	buckets       []*bucket
}

func NewBucketList(k int) *bucketList {
	b := NewBucket(k, big.NewInt(0), maxId)
	buckets := make([]*bucket, 0, k)
	buckets = append(buckets, b)
	return &bucketList{k, buckets}
}

func (bl *bucketList) find(id Id) (int, *bucket) {
	for i, b := range bl.buckets {
		if b.inRange(id) {
			return i, b
		}
	}
	return -1, nil
}

func insert(a []*bucket, i int, value *bucket) []*bucket {
	if len(a) == i { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:i+1], a[i:]...) // index < len(a)
	a[i] = value
	return a
}

func (bl *bucketList) split(i int) {
	b := bl.buckets[i]
	bucket1, bucket2 := b.split()
	bl.buckets[i] = bucket1
	bl.buckets = insert(bl.buckets, i + 1, bucket2)
}
