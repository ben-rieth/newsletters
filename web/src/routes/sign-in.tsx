import { createFileRoute } from '@tanstack/react-router';
import { SignInForm } from '#/features/auth/components/SignInForm';

export const Route = createFileRoute('/sign-in')({
  component: SignInPage,
});

const SignInPage = () => {
  return (
    <div className="flex min-h-[80vh] items-center justify-center px-4">
      <div className="w-full max-w-sm space-y-6">
        <div className="space-y-1 text-center">
          <h1 className="font-serif text-3xl font-medium tracking-tight">
            Welcome back
          </h1>
          <p className="text-sm text-muted-foreground">
            Sign in to your account
          </p>
        </div>
        <SignInForm />
      </div>
    </div>
  );
};
