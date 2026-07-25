import { createFileRoute } from '@tanstack/react-router';
import { VerifyEmailForm } from '#/features/auth/components/VerifyEmailForm';

export const Route = createFileRoute('/verify-email')({
  component: VerifyEmailPage,
});

const VerifyEmailPage = () => {
  return (
    <div className="flex min-h-[80vh] items-center justify-center px-4">
      <div className="w-full max-w-sm space-y-6">
        <div className="space-y-1 text-center">
          <h1 className="text-2xl font-bold tracking-tight">
            Check your email
          </h1>
          <p className="text-sm text-muted-foreground">
            We sent an 8-digit code to your email address.
          </p>
        </div>
        <VerifyEmailForm />
      </div>
    </div>
  );
};
