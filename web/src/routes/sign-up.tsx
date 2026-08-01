import { createFileRoute } from '@tanstack/react-router';
import { SignUpForm } from '#/features/auth/components/SignUpForm';

type SignUpSearch = {
  email?: string;
};

export const Route = createFileRoute('/sign-up')({
  component: SignUpPage,
  validateSearch: (search: Record<string, unknown>): SignUpSearch => ({
    email: typeof search.email === 'string' ? search.email : undefined,
  }),
});

const SignUpPage = () => {
  const { email } = Route.useSearch();

  return (
    <div className="flex min-h-[80vh] items-center justify-center px-4">
      <div className="w-full max-w-sm space-y-6">
        <div className="space-y-1 text-center">
          <h1 className="font-serif text-3xl font-medium tracking-tight">
            Create an account
          </h1>
          <p className="text-sm text-muted-foreground">
            Get started with your newsletters
          </p>
        </div>
        <SignUpForm initialEmail={email} />
      </div>
    </div>
  );
};
