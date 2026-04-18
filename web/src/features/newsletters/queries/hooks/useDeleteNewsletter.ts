import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { newslettersKeys } from '../newsletters'
import client from '#/api/client'
import { getErrorMessage } from '#/lib/errors'

const useDeleteNewsletter = (onSuccess?: () => void) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (newsletterId: string) => {
      const { error } = await client.DELETE('/newsletters/{newsletterId}', {
        params: { path: { newsletterId } },
      })
      if (error) {
        throw error
      }
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

export default useDeleteNewsletter
