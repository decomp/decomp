; ModuleID = 'switch.ll'
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64-unknown-linux-gnu"

; Function Attrs: nounwind uwtable
define i32 @main(i32 %argc, i8** %argv) #0 {
  switch i32 10, label %4 [
    i32 3, label %1
    i32 6, label %2
    i32 9, label %3
  ]

; <label>:1                                       ; preds = %0
  br label %2

; <label>:2                                       ; preds = %1, %0
  br label %3

; <label>:3                                       ; preds = %2, %0
  br label %4

; <label>:4                                       ; preds = %3, %0
  br label %5

; <label>:5                                       ; preds = %4
  ret i32 42
}

attributes #0 = { nounwind uwtable "less-precise-fpmad"="false" "no-frame-pointer-elim"="true" "no-frame-pointer-elim-non-leaf" "no-infs-fp-math"="false" "no-nans-fp-math"="false" "stack-protector-buffer-size"="8" "unsafe-fp-math"="false" "use-soft-float"="false" }

!llvm.ident = !{!0}

!0 = !{!"clang version 3.6.0 (tags/RELEASE_360/final)"}
