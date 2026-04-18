import { createFileRoute } from '@tanstack/react-router';
import { SignUpForm } from '#/features/auth/components/SignUpForm';

export const Route = createFileRoute('/sign-up')({
  component: SignUpPage,
});

const SignUpPage = () => {
  return (
    <div className="flex min-h-[80vh] items-center justify-center">
      <div className="w-full max-w-sm space-y-6">
        <div className="space-y-1 text-center">
          <h1 className="text-2xl font-bold tracking-tight">Create an account</h1>
          <p className="text-sm text-muted-foreground">
            Get started with your newsletters
          </p>
        </div>
        <SignUpForm />
      </div>
    </div>
  );
};
