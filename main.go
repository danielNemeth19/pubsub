package main

func (u user) doBattles(subCh <-chan move) []piece {
    fights := []piece{}
    for mv := range subCh {
        for _, piece := range u.pieces {
            if piece.location == mv.piece.location {
                fights = append(fights, piece)
            }
        }
    }
    return fights
}

func (u user) march(p piece, publishChn chan<- move) {
	m := move{
		userName: u.name,
		piece:    p,
	}
	publishChn <- m
}

type user struct {
	name   string
	pieces []piece
}

type move struct {
	userName string
	piece    piece
}

type piece struct {
	location string
	name     string
}

func doBattles(publishChn <-chan move, users []user) []piece {
	figths := []piece{}
	for mv := range publishChn {
		for _, u := range users {
			if u.name == mv.userName {
				continue
			}
			for _, piece := range u.pieces {
				if piece.location == mv.piece.location {
					figths = append(figths, piece)
				}
			}
		}
	}
	return figths
}

func distributeBattles(publishChn <-chan move, subChans []chan move) {
    for mv := range publishChn {
        for _, subCh := range subChans {
            subCh <- mv
        }
    }
}
