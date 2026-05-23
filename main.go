package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ゲームの状態を管理する構造体
type GameState struct {
	mu           sync.Mutex
	PlayerTarget int
	PlayerCount  int
	PlayerStatus string
	CompLow      int
	CompHigh     int
	CompMid      int
	CompCount    int
	CompStatus   string
}

var state = &GameState{
	PlayerTarget: rand.Intn(100) + 1,
	CompLow:      1,
	CompHigh:     100,
	CompMid:      50,
	CompCount:    1,
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// HTMLとCSS（画面のデザイン）を出力
	http.HandleFunc("/", handleHome)
	// プレイヤーモードの操作処理
	http.HandleFunc("/player", handlePlayer)
	// コンピュータモードの操作処理
	http.HandleFunc("/computer", handleComputer)

	fmt.Println("==================================================")
	fmt.Println("🎉 ゲームサーバーが起動しました！")
	fmt.Println("👉 ブラウザを開いて次のURLにアクセスしてください: http://localhost:8080")
	fmt.Println("==================================================")

	// サーバーをポート8080で起動
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

// 画面全体のレイアウト（HTML/CSS）
func handleHome(w http.ResponseWriter, r *http.Request) {
	state.mu.Lock()
	defer state.mu.Unlock()

	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="ja">
	<head>
		<meta charset="UTF-8">
		<title>High & Low アルゴリズムゲーム</title>
		<style>
			body { font-family: sans-serif; background: #f0f2f5; margin: 0; padding: 20px; color: #333; }
			.container { max-width: 800px; margin: 0 auto; display: flex; gap: 20px; }
			.card { flex: 1; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
			h1 { text-align: center; color: #1e293b; }
			h2 { color: #0f172a; border-bottom: 2px solid #e2e8f0; padding-bottom: 8px; }
			.status { background: #e0f2fe; padding: 10px; border-radius: 4px; margin: 10px 0; font-weight: bold; color: #0369a1; }
			input[type="number"] { width: 80%%; padding: 8px; font-size: 16px; margin-bottom: 10px; }
			button { background: #2563eb; color: white; border: none; padding: 10px 15px; font-size: 14px; border-radius: 4px; cursor: pointer; margin-right: 5px; }
			button:hover { background: #1d4ed8; }
			.btn-reset { background: #64748b; }
			.btn-reset:hover { background: #475569; }
		</style>
	</head>
	<body>
		<h1>High & Low アルゴリズムゲーム</h1>
		<div class="container">
			
			<!-- プレイヤーモード -->
			<div class="card">
				<h2>1. プレイヤーモード</h2>
				<p>1〜100の数字を当ててください！</p>
				<form action="/player" method="POST">
					<input type="number" name="guess" min="1" max="100" required placeholder="数字を入力">
					<br>
					<button type="submit">チェック</button>
					<button type="submit" name="action" value="reset" class="btn-reset">リセット</button>
				</form>
				<div class="status">%s</div>
				<p>試行回数: %d 回</p>
			</div>

			<!-- コンピュータモード -->
			<div class="card">
				<h2>2. コンピュータモード (二分探索)</h2>
				<p>心の中で1〜100の数字を決め、ボタンで教えてください。</p>
				<div class="status">💻 コンピュータの予想: %d</div>
				<div class="status" style="background:#f1f5f9; color:#334155;">%s</div>
				<p>試行回数: %d 回</p>
				<form action="/computer" method="POST">
					<button type="submit" name="res" value="high">もっと大きい (High)</button>
					<button type="submit" name="res" value="low">もっと小さい (Low)</button>
					<button type="submit" name="res" value="correct" style="background:#16a34a;">正解！</button>
					<hr style="border:0; border-top:1px solid #e2e8f0; margin:15px 0;">
					<button type="submit" name="res" value="reset" class="btn-reset">リセット</button>
				</form>
			</div>

		</div>
	</body>
	</html>
	`, state.PlayerStatus, state.PlayerCount, state.CompMid, state.CompStatus, state.CompCount)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// プレイヤーの入力を処理
func handlePlayer(w http.ResponseWriter, r *http.Request) {
	state.mu.Lock()
	defer state.mu.Unlock()

	if r.FormValue("action") == "reset" {
		state.PlayerTarget = rand.Intn(100) + 1
		state.PlayerCount = 0
		state.PlayerStatus = "ゲームをリセットしました。数字を入力してください。"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	guess, _ := strconv.Atoi(r.FormValue("guess"))
	state.PlayerCount++

	if guess < state.PlayerTarget {
		state.PlayerStatus = fmt.Sprintf("【%d】より、もっと【 High（大きい） 】です！", guess)
	} else if guess > state.PlayerTarget {
		state.PlayerStatus = fmt.Sprintf("【%d】より、もっと【 Low（小さい） 】です！", guess)
	} else {
		state.PlayerStatus = fmt.Sprintf("🎉 正解！【%d】でした！(%d回目)", state.PlayerTarget, state.PlayerCount)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// コンピュータ（二分探索）の入力を処理
func handleComputer(w http.ResponseWriter, r *http.Request) {
	state.mu.Lock()
	defer state.mu.Unlock()

	res := r.FormValue("res")

	switch res {
	case "high":
		state.CompLow = state.CompMid + 1
		if state.CompLow > state.CompHigh {
			state.CompStatus = "矛盾しています！嘘をついていませんか？🤔"
		} else {
			state.CompCount++
			state.CompMid = state.CompLow + (state.CompHigh-state.CompLow)/2
			state.CompStatus = "もっと大きいですね、絞り込みます。"
		}
	case "low":
		state.CompHigh = state.CompMid - 1
		if state.CompLow > state.CompHigh {
			state.CompStatus = "矛盾しています！嘘をついていませんか？🤔"
		} else {
			state.CompCount++
			state.CompMid = state.CompLow + (state.CompHigh-state.CompLow)/2
			state.CompStatus = "もっと小さいですね、絞り込みます。"
		}
	case "correct":
		state.CompStatus = fmt.Sprintf("🤖「%d回で当てました！二分探索の勝利です！」", state.CompCount)
	case "reset":
		state.CompLow = 1
		state.CompHigh = 100
		state.CompMid = 50
		state.CompCount = 1
		state.CompStatus = "新しく心の中で数字を決めてください。"
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}