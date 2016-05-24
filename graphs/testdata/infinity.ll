target triple = "x86_64-unknown-linux-gnu"

define void @main() {
	br label %1

; <label>:1                                       ; preds = %2, %0
	br label %2

; <label>:2                                       ; preds = %1
	br label %1
}
