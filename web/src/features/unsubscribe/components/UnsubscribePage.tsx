import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card';
import useUnsubscribe from '../queries/hooks/useUnsubscribe';

interface UnsubscribePageProps {
  token: string;
}

const UnsubscribePage = ({ token }: UnsubscribePageProps) => {
  const [unsubscribed, setUnsubscribed] = useState(false);
  const { mutate, isPending } = useUnsubscribe(() => setUnsubscribed(true));

  if (unsubscribed) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <Card className="w-full max-w-md text-center">
          <CardHeader>
            <CardTitle>You&apos;ve been unsubscribed</CardTitle>
            <CardDescription>
              You will no longer receive emails from this newsletter. A receipt has been sent to your email.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-[60vh] items-center justify-center">
      <Card className="w-full max-w-md text-center">
        <CardHeader>
          <CardTitle>Unsubscribe</CardTitle>
          <CardDescription>
            Click the button below to unsubscribe from this newsletter.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            variant="destructive"
            onClick={() => mutate(token)}
            disabled={isPending}
          >
            {isPending ? 'Unsubscribing...' : 'Unsubscribe'}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
};

export default UnsubscribePage;
