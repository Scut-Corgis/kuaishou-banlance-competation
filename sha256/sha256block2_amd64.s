#include "textflag.h"

// Wt = Mt; for 0 <= t <= 15
// Wt = SIGMA1(Wt-2) + SIGMA0(Wt-15) + Wt-16; for 16 <= t <= 63
//
// a = H0
// b = H1
// c = H2
// d = H3
// e = H4
// f = H5
// g = H6
// h = H7
//
// for t = 0 to 63 {
//    T1 = h + BIGSIGMA1(e) + Ch(e,f,g) + Kt + Wt
//    T2 = BIGSIGMA0(a) + Maj(a,b,c)
//    h = g
//    g = f
//    f = e
//    e = d + T1
//    d = c
//    c = b
//    b = a
//    a = T1 + T2
// }
//
// H0 = a + H0
// H1 = b + H1
// H2 = c + H2
// H3 = d + H3
// H4 = e + H4
// H5 = f + H5
// H6 = g + H6
// H7 = h + H7

// Wt = Mt; for 0 <= t <= 15
#define MSGSCHEDULE0(index) \
	MOVL	(index*4)(SI), AX; \
	BSWAPL	AX; \
	MOVL	AX, (index*4)(BP)

// Wt = SIGMA1(Wt-2) + Wt-7 + SIGMA0(Wt-15) + Wt-16; for 16 <= t <= 63
//   SIGMA0(x) = ROTR(7,x) XOR ROTR(18,x) XOR SHR(3,x)
//   SIGMA1(x) = ROTR(17,x) XOR ROTR(19,x) XOR SHR(10,x)
#define MSGSCHEDULE1(index) \
	MOVL	((index-2)*4)(BP), AX; \
	MOVL	AX, CX; \
	RORL	$17, AX; \
	MOVL	CX, DX; \
	RORL	$19, CX; \
	SHRL	$10, DX; \
	MOVL	((index-15)*4)(BP), BX; \
	XORL	CX, AX; \
	MOVL	BX, CX; \
	XORL	DX, AX; \
	RORL	$7, BX; \
	MOVL	CX, DX; \
	SHRL	$3, DX; \
	RORL	$18, CX; \
	ADDL	((index-7)*4)(BP), AX; \
	XORL	CX, BX; \
	XORL	DX, BX; \
	ADDL	((index-16)*4)(BP), BX; \
	ADDL	BX, AX; \
	MOVL	AX, ((index)*4)(BP)

// Calculate T1 in AX - uses AX, CX and DX registers.
// h is also used as an accumulator. Wt is passed in AX.
//   T1 = h + BIGSIGMA1(e) + Ch(e, f, g) + Kt + Wt
//     BIGSIGMA1(x) = ROTR(6,x) XOR ROTR(11,x) XOR ROTR(25,x)
//     Ch(x, y, z) = (x AND y) XOR (NOT x AND z)
#define SHA256T1(const, e, f, g, h) \
	ADDL	AX, h; \
	MOVL	e, AX; \
	ADDL	$const, h; \
	MOVL	e, CX; \
	RORL	$6, AX; \
	MOVL	e, DX; \
	RORL	$11, CX; \
	XORL	CX, AX; \
	MOVL	e, CX; \
	RORL	$25, DX; \
	ANDL	f, CX; \
	XORL	AX, DX; \
	MOVL	e, AX; \
	NOTL	AX; \
	ADDL	DX, h; \
	ANDL	g, AX; \
	XORL	CX, AX; \
	ADDL	h, AX

// Calculate T2 in BX - uses BX, CX, DX and DI registers.
//   T2 = BIGSIGMA0(a) + Maj(a, b, c)
//     BIGSIGMA0(x) = ROTR(2,x) XOR ROTR(13,x) XOR ROTR(22,x)
//     Maj(x, y, z) = (x AND y) XOR (x AND z) XOR (y AND z)
#define SHA256T2(a, b, c) \
	MOVL	a, DI; \
	MOVL	c, BX; \
	RORL	$2, DI; \
	MOVL	a, DX; \
	ANDL	b, BX; \
	RORL	$13, DX; \
	MOVL	a, CX; \
	ANDL	c, CX; \
	XORL	DX, DI; \
	XORL	CX, BX; \
	MOVL	a, DX; \
	MOVL	b, CX; \
	RORL	$22, DX; \
	ANDL	a, CX; \
	XORL	CX, BX; \
	XORL	DX, DI; \
	ADDL	DI, BX

// Calculate T1 and T2, then e = d + T1 and a = T1 + T2.
// The values for e and a are stored in d and h, ready for rotation.
#define SHA256ROUND(index, const, a, b, c, d, e, f, g, h) \
	SHA256T1(const, e, f, g, h); \
	SHA256T2(a, b, c); \
	MOVL	BX, h; \
	ADDL	AX, d; \
	ADDL	AX, h

#define SHA256ROUND0(index, const, a, b, c, d, e, f, g, h) \
	MSGSCHEDULE0(index); \
	SHA256ROUND(index, const, a, b, c, d, e, f, g, h)

#define SHA256ROUND1(index, const, a, b, c, d, e, f, g, h) \
	MSGSCHEDULE1(index); \
	SHA256ROUND(index, const, a, b, c, d, e, f, g, h)


// Definitions for AVX2 version

// addm (mem), reg
// Add reg to mem using reg-mem add and store
#define addm(P1, P2) \
	ADDL P2, P1; \
	MOVL P1, P2

#define XDWORD0 Y4
#define XDWORD1 Y5
#define XDWORD2 Y6
#define XDWORD3 Y7

#define XWORD0 X4
#define XWORD1 X5
#define XWORD2 X6
#define XWORD3 X7

#define XTMP0 Y0
#define XTMP1 Y1
#define XTMP2 Y2
#define XTMP3 Y3
#define XTMP4 Y8
#define XTMP5 Y11

#define XFER  Y9

#define BYTE_FLIP_MASK 	Y13 // mask to convert LE -> BE
#define X_BYTE_FLIP_MASK X13

#define NUM_BYTES DX
#define INP	DI

#define CTX SI // Beginning of digest in memory (a, b, c, ... , h)

#define a AX
#define b BX
#define c CX
#define d R8
#define e DX
#define f R9
#define g R10
#define h R11

#define old_h R11

#define TBL BP

#define SRND SI // SRND is same register as CTX

#define T1 R12

#define y0 R13
#define y1 R14
#define y2 R15
#define y3 DI

// Offsets
#define XFER_SIZE 2*64*4
#define INP_END_SIZE 8
#define INP_SIZE 8

#define _XFER 0
#define _INP_END _XFER + XFER_SIZE
#define _INP _INP_END + INP_END_SIZE
#define STACK_SIZE _INP + INP_SIZE


// 实际上就是做循环这个过程的计算，可以看到a本来要对应h，但是因为是t1+t2，所以ROUND_AND_SCHED_N_0中的h最终为t1+t2，e本来对应d，所以d最终为d+t1
// for i := 0; i < 64; i++ {
// 	S1 := rightRotate(e, 6) ^ rightRotate(e, 11) ^ rightRotate(e, 25)
// 	ch := (e & f) ^ ((^e) & g)
// 	temp1 := h + S1 + ch + k[i] + w[i]
// 	S0 := rightRotate(a, 2) ^ rightRotate(a, 13) ^ rightRotate(a, 22)
// 	maj := (a & b) ^ (a & c) ^ (b & c)
// 	temp2 := S0 + maj

// 	h = g
// 	g = f
// 	f = e
// 	e = d + temp1
// 	d = c
// 	c = b
// 	b = a
// 	a = temp1 + temp2
// }

// 实际上XDWORD0, XDWORD1, XDWORD2, XDWORD3和
#define ROUND_AND_SCHED_N_0(disp, a, b, c, d, e, f, g, h) \
	;                                     \ // #############################  RND N + 0 ############################//
	MOVL     a, y3;                       \ // y3 = a					// MAJA
	RORXL    $25, e, y0;                  \ // y0 = e >> 25				// S1A
	RORXL    $11, e, y1;                  \ // y1 = e >> 11				// S1B
	;                                     \
	ADDL     (disp + 0*4)(TBL)(SRND*1), h; \ // h = k + w + h        // disp = k + w
	ORL      c, y3;                       \ // y3 = a|c				// MAJA
	MOVL     f, y2;                       \ // y2 = f				// CH
	RORXL    $13, a, T1;                  \ // T1 = a >> 13			// S0B
	;                                     \
	XORL     y1, y0;                      \ // y0 = (e>>25) ^ (e>>11)					// S1
	XORL     g, y2;                       \ // y2 = f^g                              	// CH
	RORXL    $6, e, y1;                   \ // y1 = (e >> 6)						// S1
	;                                     \
	ANDL     e, y2;                       \ // y2 = (f^g)&e                         // CH
	XORL     y1, y0;                      \ // y0 = (e>>25) ^ (e>>11) ^ (e>>6)		// S1
	RORXL    $22, a, y1;                  \ // y1 = a >> 22							// S0A
	ADDL     h, d;                        \ // d = k + w + h + d                     	// --
	;                                     \
	ANDL     b, y3;                       \ // y3 = (a|c)&b							// MAJA
	XORL     T1, y1;                      \ // y1 = (a>>22) ^ (a>>13)				// S0
	RORXL    $2, a, T1;                   \ // T1 = (a >> 2)						// S0
	;                                     \
	XORL     g, y2;                       \ // y2 = CH = ((f^g)&e)^g				// CH
	XORL     T1, y1;                      \ // y1 = (a>>22) ^ (a>>13) ^ (a>>2)		// S0
	MOVL     a, T1;                       \ // T1 = a								// MAJB
	ANDL     c, T1;                       \ // T1 = a&c								// MAJB
	;                                     \
	ADDL     y0, y2;                      \ // y2 = S1 + CH							// --
	ORL      T1, y3;                      \ // y3 = MAJ = (a|c)&b)|(a&c)			// MAJ
	ADDL     y1, h;                       \ // h = k + w + h + S0					// --
	;                                     \
	ADDL     y2, d;                       \ // d = k + w + h + d + S1 + CH = d + t1  // --
	;                                     \
	ADDL     y2, h;                       \ // h = k + w + h + S0 + S1 + CH = t1 + S0// --
	ADDL     y3, h                        // h = t1 + S0 + MAJ                     // --

#define ROUND_AND_SCHED_N_1(disp, a, b, c, d, e, f, g, h) \
	;                                    \ // ################################### RND N + 1 ############################
	;                                    \
	MOVL    a, y3;                       \ // y3 = a                       // MAJA
	RORXL   $25, e, y0;                  \ // y0 = e >> 25					// S1A
	RORXL   $11, e, y1;                  \ // y1 = e >> 11					// S1B
	ADDL    (disp + 1*4)(TBL)(SRND*1), h; \ // h = k + w + h         		// --
	ORL     c, y3;                       \ // y3 = a|c						// MAJA
	;                                    \
	MOVL    f, y2;                       \ // y2 = f						// CH
	RORXL   $13, a, T1;                  \ // T1 = a >> 13					// S0B
	XORL    y1, y0;                      \ // y0 = (e>>25) ^ (e>>11)		// S1
	XORL    g, y2;                       \ // y2 = f^g						// CH
	;                                    \
	RORXL   $6, e, y1;                   \ // y1 = (e >> 6)				// S1
	XORL    y1, y0;                      \ // y0 = (e>>25) ^ (e>>11) ^ (e>>6)	// S1
	RORXL   $22, a, y1;                  \ // y1 = a >> 22						// S0A
	ANDL    e, y2;                       \ // y2 = (f^g)&e						// CH
	ADDL    h, d;                        \ // d = k + w + h + d				// --
	;                                    \
	ANDL    b, y3;                       \ // y3 = (a|c)&b					// MAJA
	XORL    T1, y1;                      \ // y1 = (a>>22) ^ (a>>13)		// S0
	;                                    \
	RORXL   $2, a, T1;                   \ // T1 = (a >> 2)				// S0
	XORL    g, y2;                       \ // y2 = CH = ((f^g)&e)^g		// CH
	;                                    \
	XORL    T1, y1;                      \ // y1 = (a>>22) ^ (a>>13) ^ (a>>2)		// S0
	MOVL    a, T1;                       \ // T1 = a						// MAJB
	ANDL    c, T1;                       \ // T1 = a&c						// MAJB
	ADDL    y0, y2;                      \ // y2 = S1 + CH					// --
	;                                    \
	ORL     T1, y3;                      \ // y3 = MAJ = (a|c)&b)|(a&c)             // MAJ
	ADDL    y1, h;                       \ // h = k + w + h + S0                    // --
	;                                    \
	ADDL    y2, d;                       \ // d = k + w + h + d + S1 + CH = d + t1  // --
	ADDL    y2, h;                       \ // h = k + w + h + S0 + S1 + CH = t1 + S0// --
	ADDL    y3, h;                       \ // h = t1 + S0 + MAJ                     // --
	;                                    \

#define ROUND_AND_SCHED_N_2(disp, a, b, c, d, e, f, g, h) \
	;                                    \ // ################################### RND N + 2 ############################
	;                                    \
	MOVL    a, y3;                       \ // y3 = a							// MAJA
	RORXL   $25, e, y0;                  \ // y0 = e >> 25						// S1A
	ADDL    (disp + 2*4)(TBL)(SRND*1), h; \ // h = k + w + h        			// --
	;                                    \
	RORXL   $11, e, y1;                  \ // y1 = e >> 11						// S1B
	ORL     c, y3;                       \ // y3 = a|c                         // MAJA
	MOVL    f, y2;                       \ // y2 = f                           // CH
	XORL    g, y2;                       \ // y2 = f^g                         // CH
	;                                    \
	RORXL   $13, a, T1;                  \ // T1 = a >> 13						// S0B
	XORL    y1, y0;                      \ // y0 = (e>>25) ^ (e>>11)			// S1
	ANDL    e, y2;                       \ // y2 = (f^g)&e						// CH
	;                                    \
	RORXL   $6, e, y1;                   \ // y1 = (e >> 6)					// S1
	ADDL    h, d;                        \ // d = k + w + h + d				// --
	ANDL    b, y3;                       \ // y3 = (a|c)&b						// MAJA
	;                                    \
	XORL    y1, y0;                      \ // y0 = (e>>25) ^ (e>>11) ^ (e>>6)	// S1
	RORXL   $22, a, y1;                  \ // y1 = a >> 22						// S0A
	XORL    g, y2;                       \ // y2 = CH = ((f^g)&e)^g			// CH
	;                                    \
	;                                    \
	XORL    T1, y1;                      \ // y1 = (a>>22) ^ (a>>13)		// S0
	RORXL   $2, a, T1;                   \ // T1 = (a >> 2)				// S0
	;                                    \
	XORL    T1, y1;                      \ // y1 = (a>>22) ^ (a>>13) ^ (a>>2)	// S0
	MOVL    a, T1;                       \ // T1 = a                                // MAJB
	ANDL    c, T1;                       \ // T1 = a&c                              // MAJB
	ADDL    y0, y2;                      \ // y2 = S1 + CH                          // --
	;                                    \
	ORL     T1, y3;                      \ // y3 = MAJ = (a|c)&b)|(a&c)             // MAJ
	ADDL    y1, h;                       \ // h = k + w + h + S0                    // --
	ADDL    y2, d;                       \ // d = k + w + h + d + S1 + CH = d + t1  // --
	ADDL    y2, h;                       \ // h = k + w + h + S0 + S1 + CH = t1 + S0// --
	;                                    \
	ADDL    y3, h                        // h = t1 + S0 + MAJ                     // --

#define ROUND_AND_SCHED_N_3(disp, a, b, c, d, e, f, g, h) \
	;                                    \ // ################################### RND N + 3 ############################
	;                                    \
	MOVL    a, y3;                       \ // y3 = a						// MAJA
	RORXL   $25, e, y0;                  \ // y0 = e >> 25					// S1A
	RORXL   $11, e, y1;                  \ // y1 = e >> 11					// S1B
	ADDL    (disp + 3*4)(TBL)(SRND*1), h; \ // h = k + w + h				// --
	ORL     c, y3;                       \ // y3 = a|c                     // MAJA
	;                                    \
	MOVL    f, y2;                       \ // y2 = f						// CH
	RORXL   $13, a, T1;                  \ // T1 = a >> 13					// S0B
	XORL    y1, y0;                      \ // y0 = (e>>25) ^ (e>>11)		// S1
	XORL    g, y2;                       \ // y2 = f^g						// CH
	;                                    \
	RORXL   $6, e, y1;                   \ // y1 = (e >> 6)				// S1
	ANDL    e, y2;                       \ // y2 = (f^g)&e					// CH
	ADDL    h, d;                        \ // d = k + w + h + d			// --
	ANDL    b, y3;                       \ // y3 = (a|c)&b					// MAJA
	;                                    \
	XORL    y1, y0;                      \ // y0 = (e>>25) ^ (e>>11) ^ (e>>6)	// S1
	XORL    g, y2;                       \ // y2 = CH = ((f^g)&e)^g			// CH
	;                                    \
	RORXL   $22, a, y1;                  \ // y1 = a >> 22					// S0A
	ADDL    y0, y2;                      \ // y2 = S1 + CH					// --
	;                                    \
	XORL    T1, y1;                      \ // y1 = (a>>22) ^ (a>>13)		// S0
	ADDL    y2, d;                       \ // d = k + w + h + d + S1 + CH = d + t1  // --
	;                                    \
	RORXL   $2, a, T1;                   \ // T1 = (a >> 2)				// S0
	;                                    \
	;                                    \
	XORL    T1, y1;                      \ // y1 = (a>>22) ^ (a>>13) ^ (a>>2)	// S0
	MOVL    a, T1;                       \ // T1 = a							// MAJB
	ANDL    c, T1;                       \ // T1 = a&c							// MAJB
	ORL     T1, y3;                      \ // y3 = MAJ = (a|c)&b)|(a&c)		// MAJ
	;                                    \
	ADDL    y1, h;                       \ // h = k + w + h + S0				// --
	ADDL    y2, h;                       \ // h = k + w + h + S0 + S1 + CH = t1 + S0// --
	ADDL    y3, h                        // h = t1 + S0 + MAJ				// --

#define DO_ROUND_N_0(disp, a, b, c, d, e, f, g, h, old_h) \
	;                                  \ // ################################### RND N + 0 ###########################
	MOVL  f, y2;                       \ // y2 = f					// CH
	RORXL $25, e, y0;                  \ // y0 = e >> 25				// S1A
	RORXL $11, e, y1;                  \ // y1 = e >> 11				// S1B
	XORL  g, y2;                       \ // y2 = f^g					// CH
	;                                  \
	XORL  y1, y0;                      \ // y0 = (e>>25) ^ (e>>11)	// S1
	RORXL $6, e, y1;                   \ // y1 = (e >> 6)			// S1
	ANDL  e, y2;                       \ // y2 = (f^g)&e				// CH
	;                                  \
	XORL  y1, y0;                      \ // y0 = (e>>25) ^ (e>>11) ^ (e>>6)	// S1
	RORXL $13, a, T1;                  \ // T1 = a >> 13						// S0B
	XORL  g, y2;                       \ // y2 = CH = ((f^g)&e)^g			// CH
	RORXL $22, a, y1;                  \ // y1 = a >> 22						// S0A
	MOVL  a, y3;                       \ // y3 = a							// MAJA
	;                                  \
	XORL  T1, y1;                      \ // y1 = (a>>22) ^ (a>>13)			// S0
	RORXL $2, a, T1;                   \ // T1 = (a >> 2)					// S0
	ADDL  (disp + 0*4)(TBL)(SRND*1), h; \ // h = k + w + h // --
	ORL   c, y3;                       \ // y3 = a|c							// MAJA
	;                                  \
	XORL  T1, y1;                      \ // y1 = (a>>22) ^ (a>>13) ^ (a>>2)	// S0
	MOVL  a, T1;                       \ // T1 = a							// MAJB
	ANDL  b, y3;                       \ // y3 = (a|c)&b						// MAJA
	ANDL  c, T1;                       \ // T1 = a&c							// MAJB
	ADDL  y0, y2;                      \ // y2 = S1 + CH						// --
	;                                  \
	ADDL  h, d;                        \ // d = k + w + h + d					// --
	ORL   T1, y3;                      \ // y3 = MAJ = (a|c)&b)|(a&c)			// MAJ
	ADDL  y1, h;                       \ // h = k + w + h + S0					// --
	ADDL  y2, d                        // d = k + w + h + d + S1 + CH = d + t1	// --

#define DO_ROUND_N_1(disp, a, b, c, d, e, f, g, h, old_h) \
	;                                  \ // ################################### RND N + 1 ###########################
	ADDL  y2, old_h;                   \ // h = k + w + h + S0 + S1 + CH = t1 + S0 // --
	MOVL  f, y2;                       \ // y2 = f                                // CH
	RORXL $25, e, y0;                  \ // y0 = e >> 25				// S1A
	RORXL $11, e, y1;                  \ // y1 = e >> 11				// S1B
	XORL  g, y2;                       \ // y2 = f^g                             // CH
	;                                  \
	XORL  y1, y0;                      \ // y0 = (e>>25) ^ (e>>11)				// S1
	RORXL $6, e, y1;                   \ // y1 = (e >> 6)						// S1
	ANDL  e, y2;                       \ // y2 = (f^g)&e                         // CH
	ADDL  y3, old_h;                   \ // h = t1 + S0 + MAJ                    // --
	;                                  \
	XORL  y1, y0;                      \ // y0 = (e>>25) ^ (e>>11) ^ (e>>6)		// S1
	RORXL $13, a, T1;                  \ // T1 = a >> 13							// S0B
	XORL  g, y2;                       \ // y2 = CH = ((f^g)&e)^g                // CH
	RORXL $22, a, y1;                  \ // y1 = a >> 22							// S0A
	MOVL  a, y3;                       \ // y3 = a                               // MAJA
	;                                  \
	XORL  T1, y1;                      \ // y1 = (a>>22) ^ (a>>13)				// S0
	RORXL $2, a, T1;                   \ // T1 = (a >> 2)						// S0
	ADDL  (disp + 1*4)(TBL)(SRND*1), h; \ // h = k + w + h // --
	ORL   c, y3;                       \ // y3 = a|c                             // MAJA
	;                                  \
	XORL  T1, y1;                      \ // y1 = (a>>22) ^ (a>>13) ^ (a>>2)		// S0
	MOVL  a, T1;                       \ // T1 = a                               // MAJB
	ANDL  b, y3;                       \ // y3 = (a|c)&b                         // MAJA
	ANDL  c, T1;                       \ // T1 = a&c                             // MAJB
	ADDL  y0, y2;                      \ // y2 = S1 + CH                         // --
	;                                  \
	ADDL  h, d;                        \ // d = k + w + h + d                    // --
	ORL   T1, y3;                      \ // y3 = MAJ = (a|c)&b)|(a&c)            // MAJ
	ADDL  y1, h;                       \ // h = k + w + h + S0                   // --
	;                                  \
	ADDL  y2, d                        // d = k + w + h + d + S1 + CH = d + t1 // --

#define DO_ROUND_N_2(disp, a, b, c, d, e, f, g, h, old_h) \
	;                                  \ // ################################### RND N + 2 ##############################
	ADDL  y2, old_h;                   \ // h = k + w + h + S0 + S1 + CH = t1 + S0// --
	MOVL  f, y2;                       \ // y2 = f								// CH
	RORXL $25, e, y0;                  \ // y0 = e >> 25							// S1A
	RORXL $11, e, y1;                  \ // y1 = e >> 11							// S1B
	XORL  g, y2;                       \ // y2 = f^g								// CH
	;                                  \
	XORL  y1, y0;                      \ // y0 = (e>>25) ^ (e>>11)				// S1
	RORXL $6, e, y1;                   \ // y1 = (e >> 6)						// S1
	ANDL  e, y2;                       \ // y2 = (f^g)&e							// CH
	ADDL  y3, old_h;                   \ // h = t1 + S0 + MAJ					// --
	;                                  \
	XORL  y1, y0;                      \ // y0 = (e>>25) ^ (e>>11) ^ (e>>6)		// S1
	RORXL $13, a, T1;                  \ // T1 = a >> 13							// S0B
	XORL  g, y2;                       \ // y2 = CH = ((f^g)&e)^g                // CH
	RORXL $22, a, y1;                  \ // y1 = a >> 22							// S0A
	MOVL  a, y3;                       \ // y3 = a								// MAJA
	;                                  \
	XORL  T1, y1;                      \ // y1 = (a>>22) ^ (a>>13)				// S0
	RORXL $2, a, T1;                   \ // T1 = (a >> 2)						// S0
	ADDL  (disp + 2*4)(TBL)(SRND*1), h; \ // h = k + w + h 	// --
	ORL   c, y3;                       \ // y3 = a|c								// MAJA
	;                                  \
	XORL  T1, y1;                      \ // y1 = (a>>22) ^ (a>>13) ^ (a>>2)		// S0
	MOVL  a, T1;                       \ // T1 = a								// MAJB
	ANDL  b, y3;                       \ // y3 = (a|c)&b							// MAJA
	ANDL  c, T1;                       \ // T1 = a&c								// MAJB
	ADDL  y0, y2;                      \ // y2 = S1 + CH							// --
	;                                  \
	ADDL  h, d;                        \ // d = k + w + h + d					// --
	ORL   T1, y3;                      \ // y3 = MAJ = (a|c)&b)|(a&c)			// MAJ
	ADDL  y1, h;                       \ // h = k + w + h + S0					// --
	;                                  \
	ADDL  y2, d                        // d = k + w + h + d + S1 + CH = d + t1 // --

#define DO_ROUND_N_3(disp, a, b, c, d, e, f, g, h, old_h) \
	;                                  \ // ################################### RND N + 3 ###########################
	ADDL  y2, old_h;                   \ // h = k + w + h + S0 + S1 + CH = t1 + S0// --
	MOVL  f, y2;                       \ // y2 = f								// CH
	RORXL $25, e, y0;                  \ // y0 = e >> 25							// S1A
	RORXL $11, e, y1;                  \ // y1 = e >> 11							// S1B
	XORL  g, y2;                       \ // y2 = f^g								// CH
	;                                  \
	XORL  y1, y0;                      \ // y0 = (e>>25) ^ (e>>11)				// S1
	RORXL $6, e, y1;                   \ // y1 = (e >> 6)						// S1
	ANDL  e, y2;                       \ // y2 = (f^g)&e							// CH
	ADDL  y3, old_h;                   \ // h = t1 + S0 + MAJ					// --
	;                                  \
	XORL  y1, y0;                      \ // y0 = (e>>25) ^ (e>>11) ^ (e>>6)		// S1
	RORXL $13, a, T1;                  \ // T1 = a >> 13							// S0B
	XORL  g, y2;                       \ // y2 = CH = ((f^g)&e)^g				// CH
	RORXL $22, a, y1;                  \ // y1 = a >> 22							// S0A
	MOVL  a, y3;                       \ // y3 = a								// MAJA
	;                                  \
	XORL  T1, y1;                      \ // y1 = (a>>22) ^ (a>>13)				// S0
	RORXL $2, a, T1;                   \ // T1 = (a >> 2)						// S0
	ADDL  (disp + 3*4)(TBL)(SRND*1), h; \ // h = k + w + h 	// --
	ORL   c, y3;                       \ // y3 = a|c								// MAJA
	;                                  \
	XORL  T1, y1;                      \ // y1 = (a>>22) ^ (a>>13) ^ (a>>2)		// S0
	MOVL  a, T1;                       \ // T1 = a								// MAJB
	ANDL  b, y3;                       \ // y3 = (a|c)&b							// MAJA
	ANDL  c, T1;                       \ // T1 = a&c								// MAJB
	ADDL  y0, y2;                      \ // y2 = S1 + CH							// --
	;                                  \
	ADDL  h, d;                        \ // d = k + w + h + d					// --
	ORL   T1, y3;                      \ // y3 = MAJ = (a|c)&b)|(a&c)			// MAJ
	ADDL  y1, h;                       \ // h = k + w + h + S0					// --
	;                                  \
	ADDL  y2, d;                       \ // d = k + w + h + d + S1 + CH = d + t1	// --
	;                                  \
	ADDL  y2, h;                       \ // h = k + w + h + S0 + S1 + CH = t1 + S0// --
	;                                  \
	ADDL  y3, h                        // h = t1 + S0 + MAJ					// --

TEXT ·Block2(SB), 0, $536-32

	MOVQ dig+0(FP), CTX          // d.h[8]

	// Load initial digest
	MOVL 0(CTX), a  // a = H0
	MOVL 4(CTX), b  // b = H1
	MOVL 8(CTX), c  // c = H2
	MOVL 12(CTX), d // d = H3
	MOVL 16(CTX), e // e = H4
	MOVL 20(CTX), f // f = H5
	MOVL 24(CTX), g // g = H6
	MOVL 28(CTX), h // h = H7

	MOVQ $KW<>(SB), TBL

	ADDQ $64, INP
	MOVQ INP, _INP(SP)
	XORQ SRND, SRND

avx2_loop1: // for w0 - w47
	// Do 4 rounds and scheduling

    // 将byte[0-15]组成的16字节 + [K0-K4]组成的16字节，存到栈的指定位置
    // 其实就是一次做了4个计算，即w0 + k0， w1 + k1， w2 + k2, w3 + k3，因为byte[0-3]就是对应w[0]
    // XDWORD0 对应 byte[0 - 15]

	// VMOVDQU 0*32(TBL)(SRND*1), XFER        
	// VMOVDQU XFER, (_XFER + 0*32)(SP)(SRND*1) 

	ROUND_AND_SCHED_N_0(_XFER + 0*32, a, b, c, d, e, f, g, h)
	ROUND_AND_SCHED_N_1(_XFER + 0*32, h, a, b, c, d, e, f, g)
	ROUND_AND_SCHED_N_2(_XFER + 0*32, g, h, a, b, c, d, e, f)
	ROUND_AND_SCHED_N_3(_XFER + 0*32, f, g, h, a, b, c, d, e)

	// Do 4 rounds and scheduling
	// VMOVDQU  1*32(TBL)(SRND*1), XFER
	// VMOVDQU XFER, (_XFER + 1*32)(SP)(SRND*1)
	ROUND_AND_SCHED_N_0(_XFER + 1*32, e, f, g, h, a, b, c, d)
	ROUND_AND_SCHED_N_1(_XFER + 1*32, d, e, f, g, h, a, b, c)
	ROUND_AND_SCHED_N_2(_XFER + 1*32, c, d, e, f, g, h, a, b)
	ROUND_AND_SCHED_N_3(_XFER + 1*32, b, c, d, e, f, g, h, a)

	// Do 4 rounds and scheduling
	// VMOVDQU  2*32(TBL)(SRND*1), XFER
	// VMOVDQU XFER, (_XFER + 2*32)(SP)(SRND*1)
	ROUND_AND_SCHED_N_0(_XFER + 2*32, a, b, c, d, e, f, g, h)
	ROUND_AND_SCHED_N_1(_XFER + 2*32, h, a, b, c, d, e, f, g)
	ROUND_AND_SCHED_N_2(_XFER + 2*32, g, h, a, b, c, d, e, f)
	ROUND_AND_SCHED_N_3(_XFER + 2*32, f, g, h, a, b, c, d, e)

	// Do 4 rounds and scheduling
	// VMOVDQU  3*32(TBL)(SRND*1), XFER
	// VMOVDQU XFER, (_XFER + 3*32)(SP)(SRND*1)
	ROUND_AND_SCHED_N_0(_XFER + 3*32, e, f, g, h, a, b, c, d)
	ROUND_AND_SCHED_N_1(_XFER + 3*32, d, e, f, g, h, a, b, c)
	ROUND_AND_SCHED_N_2(_XFER + 3*32, c, d, e, f, g, h, a, b)
	ROUND_AND_SCHED_N_3(_XFER + 3*32, b, c, d, e, f, g, h, a)

	ADDQ $4*32, SRND
	CMPQ SRND, $3*4*32
	JB   avx2_loop1

avx2_loop2:

	// VXORPS XDWORD0, XDWORD0, XDWORD0
	// VXORPS XDWORD1, XDWORD1, XDWORD1
	// VXORPS XDWORD2, XDWORD2, XDWORD2
	// VXORPS XDWORD3, XDWORD3, XDWORD3
	// w48 - w63 processed with no scheduling (last 16 rounds)
	// VMOVDQU 0*32(TBL)(SRND*1), XFER        
	// VMOVDQU XFER, (_XFER + 0*32)(SP)(SRND*1) 

	// VPADDD  0*32(TBL)(SRND*1), XDWORD0, XFER
	// VMOVDQU XFER, (_XFER + 0*32)(SP)(SRND*1)
	DO_ROUND_N_0(_XFER + 0*32, a, b, c, d, e, f, g, h, h)
	DO_ROUND_N_1(_XFER + 0*32, h, a, b, c, d, e, f, g, h)
	DO_ROUND_N_2(_XFER + 0*32, g, h, a, b, c, d, e, f, g)
	DO_ROUND_N_3(_XFER + 0*32, f, g, h, a, b, c, d, e, f)

	// VMOVDQU 1*32(TBL)(SRND*1), XFER        
	// VMOVDQU XFER, (_XFER + 1*32)(SP)(SRND*1) 

	// VPADDD  1*32(TBL)(SRND*1), XDWORD1, XFER
	// VMOVDQU XFER, (_XFER + 1*32)(SP)(SRND*1)
	DO_ROUND_N_0(_XFER + 1*32, e, f, g, h, a, b, c, d, e)
	DO_ROUND_N_1(_XFER + 1*32, d, e, f, g, h, a, b, c, d)
	DO_ROUND_N_2(_XFER + 1*32, c, d, e, f, g, h, a, b, c)
	DO_ROUND_N_3(_XFER + 1*32, b, c, d, e, f, g, h, a, b)

	ADDQ $2*32, SRND

	// VMOVDQU XDWORD2, XDWORD0
	// VMOVDQU XDWORD3, XDWORD1
	// 循环两次就退出
	CMPQ SRND, $4*4*32
	JB   avx2_loop2

	MOVQ dig+0(FP), CTX // d.h[8]
	MOVQ _INP(SP), INP

	addm(  0(CTX), a)
	addm(  4(CTX), b)
	addm(  8(CTX), c)
	addm( 12(CTX), d)
	addm( 16(CTX), e)
	addm( 20(CTX), f)
	addm( 24(CTX), g)
	addm( 28(CTX), h)

	// 直接把小端转大端返回，后面就不用uintput32了
	VMOVDQU (0*32)(CTX), XTMP0
	VMOVDQU flip_mask<>(SB), BYTE_FLIP_MASK

	// Apply Byte Flip Mask: LE -> BE
	VPSHUFB BYTE_FLIP_MASK, XTMP0, XTMP0
	VMOVDQU XTMP0, (0*32)(CTX)

	VZEROUPPER
	RET

// shuffle byte order from LE to BE
DATA flip_mask<>+0x00(SB)/8, $0x0405060700010203
DATA flip_mask<>+0x08(SB)/8, $0x0c0d0e0f08090a0b
DATA flip_mask<>+0x10(SB)/8, $0x0405060700010203
DATA flip_mask<>+0x18(SB)/8, $0x0c0d0e0f08090a0b
GLOBL flip_mask<>(SB), 8, $32

// shuffle xBxA -> 00BA
DATA shuff_00BA<>+0x00(SB)/8, $0x0b0a090803020100
DATA shuff_00BA<>+0x08(SB)/8, $0xFFFFFFFFFFFFFFFF
DATA shuff_00BA<>+0x10(SB)/8, $0x0b0a090803020100
DATA shuff_00BA<>+0x18(SB)/8, $0xFFFFFFFFFFFFFFFF
GLOBL shuff_00BA<>(SB), 8, $32

// shuffle xDxC -> DC00
DATA shuff_DC00<>+0x00(SB)/8, $0xFFFFFFFFFFFFFFFF
DATA shuff_DC00<>+0x08(SB)/8, $0x0b0a090803020100
DATA shuff_DC00<>+0x10(SB)/8, $0xFFFFFFFFFFFFFFFF
DATA shuff_DC00<>+0x18(SB)/8, $0x0b0a090803020100
GLOBL shuff_DC00<>(SB), 8, $32

// Round specific constants
DATA K256<>+0x00(SB)/4, $0x428a2f98 // k1
DATA K256<>+0x04(SB)/4, $0x71374491 // k2
DATA K256<>+0x08(SB)/4, $0xb5c0fbcf // k3
DATA K256<>+0x0c(SB)/4, $0xe9b5dba5 // k4
DATA K256<>+0x10(SB)/4, $0x428a2f98 // k1
DATA K256<>+0x14(SB)/4, $0x71374491 // k2
DATA K256<>+0x18(SB)/4, $0xb5c0fbcf // k3
DATA K256<>+0x1c(SB)/4, $0xe9b5dba5 // k4

DATA K256<>+0x20(SB)/4, $0x3956c25b // k5 - k8
DATA K256<>+0x24(SB)/4, $0x59f111f1
DATA K256<>+0x28(SB)/4, $0x923f82a4
DATA K256<>+0x2c(SB)/4, $0xab1c5ed5
DATA K256<>+0x30(SB)/4, $0x3956c25b
DATA K256<>+0x34(SB)/4, $0x59f111f1
DATA K256<>+0x38(SB)/4, $0x923f82a4
DATA K256<>+0x3c(SB)/4, $0xab1c5ed5

DATA K256<>+0x40(SB)/4, $0xd807aa98 // k9 - k12
DATA K256<>+0x44(SB)/4, $0x12835b01
DATA K256<>+0x48(SB)/4, $0x243185be
DATA K256<>+0x4c(SB)/4, $0x550c7dc3
DATA K256<>+0x50(SB)/4, $0xd807aa98
DATA K256<>+0x54(SB)/4, $0x12835b01
DATA K256<>+0x58(SB)/4, $0x243185be
DATA K256<>+0x5c(SB)/4, $0x550c7dc3

DATA K256<>+0x60(SB)/4, $0x72be5d74 // k13 - k16
DATA K256<>+0x64(SB)/4, $0x80deb1fe
DATA K256<>+0x68(SB)/4, $0x9bdc06a7
DATA K256<>+0x6c(SB)/4, $0xc19bf174
DATA K256<>+0x70(SB)/4, $0x72be5d74
DATA K256<>+0x74(SB)/4, $0x80deb1fe
DATA K256<>+0x78(SB)/4, $0x9bdc06a7
DATA K256<>+0x7c(SB)/4, $0xc19bf174

DATA K256<>+0x80(SB)/4, $0xe49b69c1 // k17 - k20
DATA K256<>+0x84(SB)/4, $0xefbe4786
DATA K256<>+0x88(SB)/4, $0x0fc19dc6
DATA K256<>+0x8c(SB)/4, $0x240ca1cc
DATA K256<>+0x90(SB)/4, $0xe49b69c1
DATA K256<>+0x94(SB)/4, $0xefbe4786
DATA K256<>+0x98(SB)/4, $0x0fc19dc6
DATA K256<>+0x9c(SB)/4, $0x240ca1cc

DATA K256<>+0xa0(SB)/4, $0x2de92c6f // k21 - k24
DATA K256<>+0xa4(SB)/4, $0x4a7484aa
DATA K256<>+0xa8(SB)/4, $0x5cb0a9dc
DATA K256<>+0xac(SB)/4, $0x76f988da
DATA K256<>+0xb0(SB)/4, $0x2de92c6f
DATA K256<>+0xb4(SB)/4, $0x4a7484aa
DATA K256<>+0xb8(SB)/4, $0x5cb0a9dc
DATA K256<>+0xbc(SB)/4, $0x76f988da

DATA K256<>+0xc0(SB)/4, $0x983e5152 // k25 - k28
DATA K256<>+0xc4(SB)/4, $0xa831c66d
DATA K256<>+0xc8(SB)/4, $0xb00327c8
DATA K256<>+0xcc(SB)/4, $0xbf597fc7
DATA K256<>+0xd0(SB)/4, $0x983e5152
DATA K256<>+0xd4(SB)/4, $0xa831c66d
DATA K256<>+0xd8(SB)/4, $0xb00327c8
DATA K256<>+0xdc(SB)/4, $0xbf597fc7

DATA K256<>+0xe0(SB)/4, $0xc6e00bf3 // k29 - k32
DATA K256<>+0xe4(SB)/4, $0xd5a79147
DATA K256<>+0xe8(SB)/4, $0x06ca6351
DATA K256<>+0xec(SB)/4, $0x14292967
DATA K256<>+0xf0(SB)/4, $0xc6e00bf3
DATA K256<>+0xf4(SB)/4, $0xd5a79147
DATA K256<>+0xf8(SB)/4, $0x06ca6351
DATA K256<>+0xfc(SB)/4, $0x14292967

DATA K256<>+0x100(SB)/4, $0x27b70a85
DATA K256<>+0x104(SB)/4, $0x2e1b2138
DATA K256<>+0x108(SB)/4, $0x4d2c6dfc
DATA K256<>+0x10c(SB)/4, $0x53380d13
DATA K256<>+0x110(SB)/4, $0x27b70a85
DATA K256<>+0x114(SB)/4, $0x2e1b2138
DATA K256<>+0x118(SB)/4, $0x4d2c6dfc
DATA K256<>+0x11c(SB)/4, $0x53380d13

DATA K256<>+0x120(SB)/4, $0x650a7354
DATA K256<>+0x124(SB)/4, $0x766a0abb
DATA K256<>+0x128(SB)/4, $0x81c2c92e
DATA K256<>+0x12c(SB)/4, $0x92722c85
DATA K256<>+0x130(SB)/4, $0x650a7354
DATA K256<>+0x134(SB)/4, $0x766a0abb
DATA K256<>+0x138(SB)/4, $0x81c2c92e
DATA K256<>+0x13c(SB)/4, $0x92722c85

DATA K256<>+0x140(SB)/4, $0xa2bfe8a1
DATA K256<>+0x144(SB)/4, $0xa81a664b
DATA K256<>+0x148(SB)/4, $0xc24b8b70
DATA K256<>+0x14c(SB)/4, $0xc76c51a3
DATA K256<>+0x150(SB)/4, $0xa2bfe8a1
DATA K256<>+0x154(SB)/4, $0xa81a664b
DATA K256<>+0x158(SB)/4, $0xc24b8b70
DATA K256<>+0x15c(SB)/4, $0xc76c51a3

DATA K256<>+0x160(SB)/4, $0xd192e819
DATA K256<>+0x164(SB)/4, $0xd6990624
DATA K256<>+0x168(SB)/4, $0xf40e3585
DATA K256<>+0x16c(SB)/4, $0x106aa070
DATA K256<>+0x170(SB)/4, $0xd192e819
DATA K256<>+0x174(SB)/4, $0xd6990624
DATA K256<>+0x178(SB)/4, $0xf40e3585
DATA K256<>+0x17c(SB)/4, $0x106aa070

DATA K256<>+0x180(SB)/4, $0x19a4c116
DATA K256<>+0x184(SB)/4, $0x1e376c08
DATA K256<>+0x188(SB)/4, $0x2748774c
DATA K256<>+0x18c(SB)/4, $0x34b0bcb5
DATA K256<>+0x190(SB)/4, $0x19a4c116
DATA K256<>+0x194(SB)/4, $0x1e376c08
DATA K256<>+0x198(SB)/4, $0x2748774c
DATA K256<>+0x19c(SB)/4, $0x34b0bcb5

DATA K256<>+0x1a0(SB)/4, $0x391c0cb3
DATA K256<>+0x1a4(SB)/4, $0x4ed8aa4a
DATA K256<>+0x1a8(SB)/4, $0x5b9cca4f
DATA K256<>+0x1ac(SB)/4, $0x682e6ff3
DATA K256<>+0x1b0(SB)/4, $0x391c0cb3
DATA K256<>+0x1b4(SB)/4, $0x4ed8aa4a
DATA K256<>+0x1b8(SB)/4, $0x5b9cca4f
DATA K256<>+0x1bc(SB)/4, $0x682e6ff3

DATA K256<>+0x1c0(SB)/4, $0x748f82ee
DATA K256<>+0x1c4(SB)/4, $0x78a5636f
DATA K256<>+0x1c8(SB)/4, $0x84c87814
DATA K256<>+0x1cc(SB)/4, $0x8cc70208
DATA K256<>+0x1d0(SB)/4, $0x748f82ee
DATA K256<>+0x1d4(SB)/4, $0x78a5636f
DATA K256<>+0x1d8(SB)/4, $0x84c87814
DATA K256<>+0x1dc(SB)/4, $0x8cc70208

DATA K256<>+0x1e0(SB)/4, $0x90befffa
DATA K256<>+0x1e4(SB)/4, $0xa4506ceb
DATA K256<>+0x1e8(SB)/4, $0xbef9a3f7
DATA K256<>+0x1ec(SB)/4, $0xc67178f2
DATA K256<>+0x1f0(SB)/4, $0x90befffa
DATA K256<>+0x1f4(SB)/4, $0xa4506ceb
DATA K256<>+0x1f8(SB)/4, $0xbef9a3f7
DATA K256<>+0x1fc(SB)/4, $0xc67178f2

GLOBL K256<>(SB), (NOPTR + RODATA), $512

// 以下为k+w的值
DATA KW<>+0x000(SB)/4, $0xc28a2f98
DATA KW<>+0x004(SB)/4, $0x71374491
DATA KW<>+0x008(SB)/4, $0xb5c0fbcf
DATA KW<>+0x00c(SB)/4, $0xe9b5dba5
DATA KW<>+0x010(SB)/4, $0xc28a2f98
DATA KW<>+0x014(SB)/4, $0x71374491
DATA KW<>+0x018(SB)/4, $0xb5c0fbcf
DATA KW<>+0x01c(SB)/4, $0xe9b5dba5
DATA KW<>+0x020(SB)/4, $0x3956c25b
DATA KW<>+0x024(SB)/4, $0x59f111f1
DATA KW<>+0x028(SB)/4, $0x923f82a4
DATA KW<>+0x02c(SB)/4, $0xab1c5ed5
DATA KW<>+0x030(SB)/4, $0x3956c25b
DATA KW<>+0x034(SB)/4, $0x59f111f1
DATA KW<>+0x038(SB)/4, $0x923f82a4
DATA KW<>+0x03c(SB)/4, $0xab1c5ed5
DATA KW<>+0x040(SB)/4, $0xd807aa98
DATA KW<>+0x044(SB)/4, $0x12835b01
DATA KW<>+0x048(SB)/4, $0x243185be
DATA KW<>+0x04c(SB)/4, $0x550c7dc3
DATA KW<>+0x050(SB)/4, $0xd807aa98
DATA KW<>+0x054(SB)/4, $0x12835b01
DATA KW<>+0x058(SB)/4, $0x243185be
DATA KW<>+0x05c(SB)/4, $0x550c7dc3
DATA KW<>+0x060(SB)/4, $0x72be5d74
DATA KW<>+0x064(SB)/4, $0x80deb1fe
DATA KW<>+0x068(SB)/4, $0x9bdc06a7
DATA KW<>+0x06c(SB)/4, $0xc19bf374
DATA KW<>+0x070(SB)/4, $0x72be5d74
DATA KW<>+0x074(SB)/4, $0x80deb1fe
DATA KW<>+0x078(SB)/4, $0x9bdc06a7
DATA KW<>+0x07c(SB)/4, $0xc19bf374
DATA KW<>+0x080(SB)/4, $0x649b69c1
DATA KW<>+0x084(SB)/4, $0xf0fe4786
DATA KW<>+0x088(SB)/4, $0x0fe1edc6
DATA KW<>+0x08c(SB)/4, $0x240cf254
DATA KW<>+0x090(SB)/4, $0x649b69c1
DATA KW<>+0x094(SB)/4, $0xf0fe4786
DATA KW<>+0x098(SB)/4, $0x0fe1edc6
DATA KW<>+0x09c(SB)/4, $0x240cf254
DATA KW<>+0x0a0(SB)/4, $0x4fe9346f
DATA KW<>+0x0a4(SB)/4, $0x6cc984be
DATA KW<>+0x0a8(SB)/4, $0x61b9411e
DATA KW<>+0x0ac(SB)/4, $0x16f988fa
DATA KW<>+0x0b0(SB)/4, $0x4fe9346f
DATA KW<>+0x0b4(SB)/4, $0x6cc984be
DATA KW<>+0x0b8(SB)/4, $0x61b9411e
DATA KW<>+0x0bc(SB)/4, $0x16f988fa
DATA KW<>+0x0c0(SB)/4, $0xf2c65152
DATA KW<>+0x0c4(SB)/4, $0xa88e5a6d
DATA KW<>+0x0c8(SB)/4, $0xb019fc65
DATA KW<>+0x0cc(SB)/4, $0xb9d99ec7
DATA KW<>+0x0d0(SB)/4, $0xf2c65152
DATA KW<>+0x0d4(SB)/4, $0xa88e5a6d
DATA KW<>+0x0d8(SB)/4, $0xb019fc65
DATA KW<>+0x0dc(SB)/4, $0xb9d99ec7
DATA KW<>+0x0e0(SB)/4, $0x9a1231c3
DATA KW<>+0x0e4(SB)/4, $0xe70eeaa0
DATA KW<>+0x0e8(SB)/4, $0xfdb1232b
DATA KW<>+0x0ec(SB)/4, $0xc7353eb0
DATA KW<>+0x0f0(SB)/4, $0x9a1231c3
DATA KW<>+0x0f4(SB)/4, $0xe70eeaa0
DATA KW<>+0x0f8(SB)/4, $0xfdb1232b
DATA KW<>+0x0fc(SB)/4, $0xc7353eb0
DATA KW<>+0x100(SB)/4, $0x3069bad5
DATA KW<>+0x104(SB)/4, $0xcb976d5f
DATA KW<>+0x108(SB)/4, $0x5a0f118f
DATA KW<>+0x10c(SB)/4, $0xdc1eeefd
DATA KW<>+0x110(SB)/4, $0x3069bad5
DATA KW<>+0x114(SB)/4, $0xcb976d5f
DATA KW<>+0x118(SB)/4, $0x5a0f118f
DATA KW<>+0x11c(SB)/4, $0xdc1eeefd
DATA KW<>+0x120(SB)/4, $0x0a35b689
DATA KW<>+0x124(SB)/4, $0xde0b7a04
DATA KW<>+0x128(SB)/4, $0x58f4ca9d
DATA KW<>+0x12c(SB)/4, $0xe15d5b16
DATA KW<>+0x130(SB)/4, $0x0a35b689
DATA KW<>+0x134(SB)/4, $0xde0b7a04
DATA KW<>+0x138(SB)/4, $0x58f4ca9d
DATA KW<>+0x13c(SB)/4, $0xe15d5b16
DATA KW<>+0x140(SB)/4, $0x007f3e86
DATA KW<>+0x144(SB)/4, $0x37088980
DATA KW<>+0x148(SB)/4, $0xa507ea32
DATA KW<>+0x14c(SB)/4, $0x6fab9537
DATA KW<>+0x150(SB)/4, $0x007f3e86
DATA KW<>+0x154(SB)/4, $0x37088980
DATA KW<>+0x158(SB)/4, $0xa507ea32
DATA KW<>+0x15c(SB)/4, $0x6fab9537
DATA KW<>+0x160(SB)/4, $0x17406110
DATA KW<>+0x164(SB)/4, $0x0d8cd6f1
DATA KW<>+0x168(SB)/4, $0xcdaa3b6d
DATA KW<>+0x16c(SB)/4, $0xc0bbbe37
DATA KW<>+0x170(SB)/4, $0x17406110
DATA KW<>+0x174(SB)/4, $0x0d8cd6f1
DATA KW<>+0x178(SB)/4, $0xcdaa3b6d
DATA KW<>+0x17c(SB)/4, $0xc0bbbe37
DATA KW<>+0x180(SB)/4, $0x83613bda
DATA KW<>+0x184(SB)/4, $0xdb48a363
DATA KW<>+0x188(SB)/4, $0x0b02e931
DATA KW<>+0x18c(SB)/4, $0x6fd15ca7
DATA KW<>+0x190(SB)/4, $0x83613bda
DATA KW<>+0x194(SB)/4, $0xdb48a363
DATA KW<>+0x198(SB)/4, $0x0b02e931
DATA KW<>+0x19c(SB)/4, $0x6fd15ca7
DATA KW<>+0x1a0(SB)/4, $0x521afaca
DATA KW<>+0x1a4(SB)/4, $0x31338431
DATA KW<>+0x1a8(SB)/4, $0x6ed41a95
DATA KW<>+0x1ac(SB)/4, $0x6d437890
DATA KW<>+0x1b0(SB)/4, $0x521afaca
DATA KW<>+0x1b4(SB)/4, $0x31338431
DATA KW<>+0x1b8(SB)/4, $0x6ed41a95
DATA KW<>+0x1bc(SB)/4, $0x6d437890
DATA KW<>+0x1c0(SB)/4, $0xc39c91f2
DATA KW<>+0x1c4(SB)/4, $0x9eccabbd
DATA KW<>+0x1c8(SB)/4, $0xb5c9a0e6
DATA KW<>+0x1cc(SB)/4, $0x532fb63c
DATA KW<>+0x1d0(SB)/4, $0xc39c91f2
DATA KW<>+0x1d4(SB)/4, $0x9eccabbd
DATA KW<>+0x1d8(SB)/4, $0xb5c9a0e6
DATA KW<>+0x1dc(SB)/4, $0x532fb63c
DATA KW<>+0x1e0(SB)/4, $0xd2c741c6
DATA KW<>+0x1e4(SB)/4, $0x07237ea3
DATA KW<>+0x1e8(SB)/4, $0xa4954b68
DATA KW<>+0x1ec(SB)/4, $0x4c191d76
DATA KW<>+0x1f0(SB)/4, $0xd2c741c6
DATA KW<>+0x1f4(SB)/4, $0x07237ea3
DATA KW<>+0x1f8(SB)/4, $0xa4954b68
DATA KW<>+0x1fc(SB)/4, $0x4c191d76

GLOBL KW<>(SB), (NOPTR + RODATA), $512
