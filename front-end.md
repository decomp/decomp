# Front-end

Translate machine code (e.g. x86 assembly) to LLVM IR.

For a (perhaps biased) comparison of machine code to LLVM IR lifters, see https://github.com/trailofbits/mcsema#comparison-with-other-machine-code-to-llvm-bitcode-lifters.

## Anvill

https://github.com/lifting-bits/anvill

Supported:
* x86 -> LLVM IR
* x86_64 -> LLVM IR
* Aarch64 -> LLVM IR

## Ghidra-to-LLVM

https://github.com/toor-de-force/Ghidra-to-LLVM

Supported:
* [Ghidra](https://github.com/NationalSecurityAgency/ghidra) [P-code](https://ghidra.re/courses/languages/html/pcoderef.html) -> LLVM IR

## Miasm

https://github.com/cea-sec/miasm

Supported:
* ARM -> Miasm IR
* Aarch64 -> Miasm IR
* MEP (big endian) -> Miasm IR
* PowerPC (32-bit big-endian) -> Miasm IR
* MIPS (32-bit) -> Miasm IR
* MSP430 -> Miasm IR
* x86 -> Miasm IR
* x86-64 -> Miasm IR

Miasm IR -> LLVM IR ([#904](https://github.com/cea-sec/miasm/pull/904))

## WAVM

https://github.com/WAVM/WAVM

Supported:
* WebAssembly -> LLVM IR

## RetDec

https://github.com/avast-tl/retdec

Supported:
* x86 -> LLVM IR
* ARM -> LLVM IR
* MIPS -> LLVM IR
* PIC32 -> LLVM IR
* PowerPC -> LLVM IR

## llvm-mctoll

https://github.com/Microsoft/llvm-mctoll

Supported:
* x86_64 -> LLVM IR
* ARM -> LLVM IR

## reopt

https://github.com/GaloisInc/reopt

Supported:
* x86_64 -> LLVM IR

## rev.ng

https://github.com/revng/revamb

Supported:
* x86_64 -> LLVM IR
* ARM -> LLVM IR
* MIPS -> LLVM IR

## MC-Semantics

https://github.com/lifting-bits/mcsema

Supported:
* x86 -> LLVM IR
* x86_64 -> LLVM IR
* Aarch64 -> LLVM IR

## Remill

https://github.com/lifting-bits/remill

Supported:
* x86 -> LLVM IR
* x86_64 -> LLVM IR
* Aarch64 -> LLVM IR

## bin2llvm

https://github.com/cojocar/bin2llvm

Supported:
* ARM -> LLVM IR

## fcd

https://github.com/zneak/fcd

Supported:
* x86_64 -> LLVM IR

## Dagger

https://github.com/repzret/dagger

Supported:
* x86 -> LLVM IR

Future:
* ARM -> LLVM IR

## RevGen

https://github.com/S2E/tools

Supported:

* x86 -> LLVM IR

## Clang

https://clang.llvm.org/

Supported:
* C -> LLVM IR
* C++ -> LLVM IR

## Fracture

https://github.com/draperlaboratory/fracture

Future:
* x86 -> LLVM IR
* MIPS -> LLVM IR
* PowerPC -> LLVM IR

## libbeauty

https://github.com/jcdutton/libbeauty

Future:
* x86 -> LLVM IR
* x86_64 -> LLVM IR

## OpenREIL

https://github.com/Cr4sh/openreil

Supported:
* x86 -> REIL
* ARM -> REIL

Future:
* x86_64 -> REIL
* REIL -> LLVM IR
