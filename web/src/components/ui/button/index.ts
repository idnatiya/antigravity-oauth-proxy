import { type VariantProps, cva } from 'class-variance-authority'

export { default as Button } from './Button.vue'

export const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-lg text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 cursor-pointer shadow-xs',
  {
    variants: {
      variant: {
        default: 'bg-blue-600 text-white shadow-sm hover:bg-blue-500',
        destructive: 'bg-red-600 text-white shadow-sm hover:bg-red-500',
        outline: 'border border-[#2c2e36] bg-[#202227] hover:bg-[#282a32] text-zinc-300 hover:text-white',
        secondary: 'bg-[#282a32] text-zinc-200 hover:bg-[#32353f]',
        ghost: 'hover:bg-[#282a32] hover:text-zinc-200 text-zinc-400',
        link: 'text-blue-400 underline-offset-4 hover:underline',
      },
      size: {
        default: 'h-9 px-4 py-2',
        sm: 'h-8 px-3 text-xs',
        lg: 'h-10 px-6 text-sm',
        icon: 'h-8 w-8 p-0',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
)

export type ButtonVariants = VariantProps<typeof buttonVariants>
