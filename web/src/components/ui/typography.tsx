import { cn } from '#/lib/utils'
import type { ReactNode } from 'react'

interface TypographyProps {
  children: ReactNode[] | ReactNode
  className?: string
}

const H1 = ({ children, className = '' }: TypographyProps) => {
  return (
    <h1
      className={cn(
        'scroll-m-20 text-4xl font-extrabold tracking-tight text-balance',
        className,
      )}
    >
      {children}
    </h1>
  )
}

const H2 = ({ children, className = '' }: TypographyProps) => {
  return (
    <h2
      className={cn(
        'scroll-m-20 border-b pb-2 text-3xl font-semibold tracking-tight first:mt-0',
        className,
      )}
    >
      {children}
    </h2>
  )
}

export { H1, H2 }
