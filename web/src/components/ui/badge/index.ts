import { type VariantProps, cva } from 'class-variance-authority'

export { default as Badge } from './Badge.vue'

export const badgeVariants = cva(
  'inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[10px] font-medium transition-colors focus:outline-none border',
  {
    variants: {
      variant: {
        default: 'border-blue-500/25 bg-blue-500/10 text-blue-400',
        secondary: 'border-zinc-700/60 bg-zinc-800/80 text-zinc-300',
        destructive: 'border-red-500/20 bg-red-500/10 text-red-400',
        outline: 'border-[#2c2e36] text-zinc-400 bg-transparent',
        success: 'border-emerald-500/20 bg-emerald-500/10 text-emerald-400',
        warning: 'border-amber-500/20 bg-amber-500/10 text-amber-400',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
)

export type BadgeVariants = VariantProps<typeof badgeVariants>
