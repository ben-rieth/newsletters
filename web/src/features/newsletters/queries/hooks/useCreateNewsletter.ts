import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { newslettersKeys } from '../newsletters'
import type { components } from '#/api/schema'
import client from '#/api/client'
import { getErrorMessage } from '#/lib/errors'

const useCreateNewsletter = (onSuccess?: () => void) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (
      newsletter: components['schemas']['SubmittableNewsletterFields'],
    ) => {
      await client.POST('/newsletters', { body: newsletter })
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: newslettersKeys.all })
      onSuccess?.()
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export default useCreateNewsletter
