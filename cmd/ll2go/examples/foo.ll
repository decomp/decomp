; ModuleID = 'foo.ll'
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64-unknown-linux-gnu"

; Function Attrs: nounwind uwtable
define i32 @main(i32 %argc, i8** %argv) #0 {
  br label %1

; <label>:1                                       ; preds = %9, %0
  %i.0 = phi i32 [ 0, %0 ], [ %10, %9 ]
  %x.0 = phi i32 [ 0, %0 ], [ %x.1, %9 ]
  %2 = icmp slt i32 %i.0, 10
  br i1 %2, label %3, label %11

; <label>:3                                       ; preds = %1
  %4 = icmp slt i32 %x.0, 100
  br i1 %4, label %5, label %8

; <label>:5                                       ; preds = %3
  %6 = mul nsw i32 3, %i.0
  %7 = add nsw i32 %x.0, %6
  br label %8

; <label>:8                                       ; preds = %5, %3
  %x.1 = phi i32 [ %7, %5 ], [ %x.0, %3 ]
  br label %9

; <label>:9                                       ; preds = %8
  %10 = add nsw i32 %i.0, 1
  br label %1

; <label>:11                                      ; preds = %1
  ret i32 %x.0
}

attributes #0 = { nounwind uwtable "less-precise-fpmad"="false" "no-frame-pointer-elim"="true" "no-frame-pointer-elim-non-leaf" "no-infs-fp-math"="false" "no-nans-fp-math"="false" "stack-protector-buffer-size"="8" "unsafe-fp-math"="false" "use-soft-float"="false" }

!llvm.ident = !{!0}

!0 = !{!"clang version 3.6.0 (tags/RELEASE_360/final)"}
