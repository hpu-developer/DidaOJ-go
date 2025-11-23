package bot

// GomokuMove 表示五子棋的一步棋
type GomokuMove struct {
	X    int   // X坐标
	Y    int   // Y坐标
	Side int   // 0表示黑棋，1表示白棋
	Err  error // 错误信息
}

// GomokuBoard 表示五子棋棋盘
type GomokuBoard struct {
	Board [][]int // 0表示空，1表示黑棋，2表示白棋
	Size  int     // 棋盘大小
}

// NewGomokuBoard 创建一个新的五子棋棋盘
func NewGomokuBoard(size int) *GomokuBoard {
	board := make([][]int, size)
	for i := range board {
		board[i] = make([]int, size)
	}
	return &GomokuBoard{
		Board: board,
		Size:  size,
	}
}

// MakeMove 在棋盘上落子
func (b *GomokuBoard) MakeMove(x, y, side int) bool {
	// 检查坐标是否有效
	if x < 0 || x >= b.Size || y < 0 || y >= b.Size {
		return false
	}
	// 检查是否已有棋子
	if b.Board[y][x] != 0 {
		return false
	}
	// 落子：黑棋为1，白棋为2
	piece := 1
	if side == 1 {
		piece = 2
	}
	b.Board[y][x] = piece
	return true
}

// CheckWin 检查是否获胜
func (b *GomokuBoard) CheckWin(x, y int) bool {
	piece := b.Board[y][x]
	if piece == 0 {
		return false
	}

	// 检查四个方向：水平、垂直、两个对角线
	directions := [][2]int{
		{1, 0},  // 水平
		{0, 1},  // 垂直
		{1, 1},  // 对角线
		{1, -1}, // 反对角线
	}

	for _, dir := range directions {
		count := 1 // 当前位置已经有一个棋子

		// 正方向
		for i := 1; i < 5; i++ {
			nx, ny := x+i*dir[0], y+i*dir[1]
			if nx >= 0 && nx < b.Size && ny >= 0 && ny < b.Size && b.Board[ny][nx] == piece {
				count++
			} else {
				break
			}
		}

		// 反方向
		for i := 1; i < 5; i++ {
			nx, ny := x-i*dir[0], y-i*dir[1]
			if nx >= 0 && nx < b.Size && ny >= 0 && ny < b.Size && b.Board[ny][nx] == piece {
				count++
			} else {
				break
			}
		}

		// 五子连珠
		if count >= 5 {
			return true
		}
	}

	return false
}
