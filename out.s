.section .bss
    .align 4
    .global buffer
buffer:
    .zero 32
.section .data
    .align 4
newline:
    .byte 10
.section .text
.global _start
_start:
    stp x29, x30, [sp, #-16]!
    mov x29, sp
    // PUSH operation
    mov x0, #34
    str x0, [sp, #-16]!
    // PUSH operation
    mov x0, #35
    str x0, [sp, #-16]!
    // PLUS operation
    ldr x1, [sp], #16
    ldr x0, [sp], #16
    add x0, x0, x1
    str x0, [sp, #-16]!
    // DUMP operation
    ldr x0, [sp], #16
    bl dump
    // PUSH operation
    mov x0, #500
    str x0, [sp, #-16]!
    // PUSH operation
    mov x0, #80
    str x0, [sp, #-16]!
    // MINUS operation
    ldr x1, [sp], #16
    ldr x0, [sp], #16
    sub x0, x0, x1
    str x0, [sp, #-16]!
    // DUMP operation
    ldr x0, [sp], #16
    bl dump
    mov x0, #0
    mov x8, #93
    svc #0

// Dump function
dump:
    stp x29, x30, [sp, #-16]!
    mov x29, sp
    stp x19, x20, [sp, #-16]!
    mov x19, x0
    adrp x0, buffer
    add x0, x0, :lo12:buffer
    mov x20, x0
    mov x1, #31
    mov x2, #10
    cmp x19, #0
    bne 1f
    mov w3, #48
    strb w3, [x20, x1]
    sub x1, x1, #1
    b 2f
1:
    cbz x19, 2f
    udiv x3, x19, x2
    msub x4, x3, x2, x19
    add w4, w4, #48
    strb w4, [x20, x1]
    sub x1, x1, #1
    mov x19, x3
    b 1b
2:
    add x1, x1, #1
    add x2, x20, #31
    sub x2, x2, x1
    add x1, x20, x1
    mov x0, #1
    mov x8, #64
    svc #0
    mov x0, #1
    adrp x1, newline
    add x1, x1, :lo12:newline
    mov x2, #1
    mov x8, #64
    svc #0
    ldp x19, x20, [sp], #16
    ldp x29, x30, [sp], #16
    ret
