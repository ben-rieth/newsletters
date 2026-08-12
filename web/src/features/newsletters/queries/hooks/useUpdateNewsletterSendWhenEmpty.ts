import client from '#/api/client';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { newslettersKeys } from '../newsletters';
import type { Newsletter } from '../newsletters';
import { getErrorMessage } from '#/lib/errors';

const useUpdateNewsletterSendWhenEmpty = (
  onSuccess?: (sendWhenEmpty: boolean) => void,
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      newsletterId,
      sendWhenEmpty,
    }: {
      newsletterId: string;
      sendWhenEmpty: boolean;
    }) => {
      const { error } = await client.PATCH(
        '/newsletter/{newsletterId}/send-when-empty',
        {
          params: { path: { newsletterId } },
          body: { sendWhenEmpty },
        },
      );
      if (error) {
        throw error;
      }
    },
    onMutate: async ({ newsletterId, sendWhenEmpty }) => {
      await queryClient.cancelQueries({ queryKey: newslettersKeys.all });

      const previousNewsletter = queryClient.getQueryData<Newsletter>(
        newslettersKeys.detail(newsletterId),
      );
      const previousNewsletters = queryClient.getQueryData<Newsletter[]>(
        newslettersKeys.all,
      );

      queryClient.setQueryData<Newsletter>(
        newslettersKeys.detail(newsletterId),
        (newsletter) => newsletter && { ...newsletter, sendWhenEmpty },
      );
      queryClient.setQueryData<Newsletter[]>(
        newslettersKeys.all,
        (newsletters) =>
          newsletters?.map((newsletter) =>
            newsletter.id === newsletterId
              ? { ...newsletter, sendWhenEmpty }
              : newsletter,
          ),
      );

      return { previousNewsletter, previousNewsletters };
    },
    onSuccess: (_data, { sendWhenEmpty }) => {
      onSuccess?.(sendWhenEmpty);
    },
    onError: (error, { newsletterId }, context) => {
      if (context?.previousNewsletter) {
        queryClient.setQueryData(
          newslettersKeys.detail(newsletterId),
          context.previousNewsletter,
        );
      }
      if (context?.previousNewsletters) {
        queryClient.setQueryData(
          newslettersKeys.all,
          context.previousNewsletters,
        );
      }
      toast.error(getErrorMessage(error));
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: newslettersKeys.all });
    },
  });
};

export default useUpdateNewsletterSendWhenEmpty;
