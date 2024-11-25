// +build amd64,!gccgo,!appengine

#include "textflag.h"


DATA encodeMask<>+0x00(SB)/8, $0x0f0f0f0f0f0f0f0f
DATA encodeMask<>+0x08(SB)/8, $0x0f0f0f0f0f0f0f0f
DATA encodeMask<>+0x10(SB)/8, $0x0f0f0f0f0f0f0f0f
DATA encodeMask<>+0x18(SB)/8, $0x0f0f0f0f0f0f0f0f

GLOBL encodeMask<>(SB),RODATA,$32

TEXT ·EncodeAVX(SB),NOSPLIT,$0
    MOVQ    dst+0(FP), DI
    MOVQ    src+8(FP), SI
    MOVQ    len+16(FP), BX
    MOVQ    alpha+24(FP), DX
    VMOVDQU (DX), Y15          // 加载字母表

loop:

	VMOVDQU -32(SI)(BX*1), Y0
	VPAND encodeMask<>(SB), Y0, Y1
	VPSRLW $4, Y0, Y0
	VPAND encodeMask<>(SB), Y0, Y0
	VPUNPCKHBW Y1, Y0, Y3
	VPUNPCKLBW Y1, Y0, Y0

	VPSHUFB Y3, Y15, Y3
	VPSHUFB Y0, Y15, Y0

	VPERM2I128 $0x20, Y3, Y0, Y5
	VPERM2I128 $0x31, Y3, Y0, Y6

	VMOVDQU Y5, 0(DI)
	VMOVDQU Y6, 32(DI)

ret:
    RET
