package raft

import pb "github.com/Reid00/go-raft/raftpb"

func IsEmptySnap(sp pb.Snapshot) bool {
	return sp.Metadata.Index == 0
}
