package raft

import (
	pb "github.com/Reid00/go-raft/raftpb"
)

// unstable.entries[i] has raft log position i+unstable.offset.
// Note that unstable.offset may be less than the highest log
// position in storage; this means that the next write to storage
// might need to truncate the log before persisting unstable.entries.
type unstable struct {
	// the incoming unstable snapshot, if any.
	snapshot *pb.Snapshot

	// all entries that have not yet been written to storage.
	// entries 和 snapshot 两个只存在一个
	entries []pb.Entry
	// entries 中的第一个元素的Index
	offset uint64

	logger Logger
}

// maybeFirstIndex returns the index of the first possible entry in entries
// if it has a snapshot.
// 获取的是相对整个raftLog的first index, 当无法得知时，返回的第二个参数时false
func (u *unstable) maybeFirstIndex() (uint64, bool) {
	if u.snapshot != nil {
		return u.snapshot.Metadata.Index + 1, true
	}
	return 0, false
}

// maybeLastIndex returns the last index if it has at least one
// unstable entry or snapshot.
func (u *unstable) maybeLastIndex() (uint64, bool) {
	if l := len(u.entries); l != 0 {
		return u.offset + uint64(l) - 1, true
	}
	if u.snapshot != nil {
		return u.snapshot.Metadata.Index, true
	}
	return 0, false
}

// maybeTerm returns the term of the entry at index i,
// if there is any.
func (u *unstable) maybeTerm(i uint64) (uint64, bool) {
	// i 在snapshot 中
	if i < u.offset {
		if u.snapshot == nil {
			return 0, false
		}
		if u.snapshot.Metadata.Index == i {
			return u.snapshot.Metadata.Term, true
		}
		return 0, false
	}

	last, ok := u.maybeLastIndex()
	if !ok {
		return 0, false
	}
	if i > last {
		return 0, false
	}

	return u.entries[i-u.offset].Term, true
}

// stableTo stable to snapshot before i index and t term
// stableTo 该函数传入一个索引号i和任期号t，表示应用层已经将这个索引之前的数据进行持久化了，
// 此时unstable要做的事情就是在自己的数据中查询，只有在满足任期号相同以及i大于等于offset的情况下，
// 可以将entries中的数据进行缩容，将i之前的数据删除。
func (u *unstable) stableTo(i, t uint64) {
	gt, ok := u.maybeTerm(i)
	if !ok {
		return
	}

	// if i < offset, term is matched with the snapshot
	// only update the unstable entries if term is matched with
	// an unstable entry.
	if gt == t && i >= u.offset { // 如果term 和i 是匹配的，并且i 在entries 中，则进行enties 的合并
		u.entries = u.entries[i+1-u.offset:]
		u.offset = i + 1
	}
}

// stableSnapTo 该函数传入一个索引i，用于告诉unstable，索引i对应的快照数据已经被应用层持久化了，
// 如果这个索引与当前快照数据对应的上，那么快照数据就可以被置空了。
func (u *unstable) stableSnapTo(i uint64) {
	if u.snapshot != nil && u.snapshot.Metadata.Index == i {
		u.snapshot = nil
	}
}

// restore 从快照数据中恢复，此时unstable将保存快照数据，
// 同时将offset成员设置成这个快照数据索引的下一位。
func (u *unstable) restore(s pb.Snapshot) {
	u.offset = s.Metadata.Index + 1
	u.entries = nil
	u.snapshot = &s
}

func (u *unstable) truncateAndAppend(ents []pb.Entry) {
	after := ents[0].Index

	switch {
	case after == u.offset+uint64(len(u.entries)):
		// after is the next index in the u.entries
		// directly append
		u.entries = append(u.entries, ents...)
	case after <= u.offset:
		u.logger.Infof("replace the unstable entries from index %d", after)
		// The log is being truncated to before our current offset
		// portion, so set the offset and replace the entries
		u.offset = after
		u.entries = ents
	default:
		// truncate to after and copy to u.entries then append
		u.logger.Infof("truncate the unstable entries before index %d", after)
		u.entries = append([]pb.Entry{}, u.slice(u.offset, after)...)
		u.entries = append(u.entries, ents...)
	}
}

func (u *unstable) slice(lo, hi uint64) []pb.Entry {
	u.mustCheckOutOfBounds(lo, hi)
	return u.entries[lo-u.offset : hi-u.offset]
}

func (u *unstable) mustCheckOutOfBounds(lo, hi uint64) {
	if lo > hi {
		u.logger.Panicf("invalid unstable.slice %d > %d", lo, hi)
	}

	upper := u.offset + uint64(len(u.entries))
	if lo < u.offset || hi > upper {
		u.logger.Panicf("unstable.slice(%d, %d) out of boud [%d, %d]", lo, hi, u.offset, upper)
	}
}
