import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card';
import { Button } from '#/components/ui/button';
import { H2 } from '#/components/ui/typography';
import { UpdateEmailForm } from '#/features/auth/components/UpdateEmailForm';
import { UpdatePasswordForm } from '#/features/auth/components/UpdatePasswordForm';
import { DeleteAccountForm } from '#/features/auth/components/DeleteAccountForm';
import { userOptions } from '#/features/auth/queries/user';
import useLogout from '#/features/auth/queries/hooks/useLogout';

const ProfilePage = () => {
  const { data: user } = useSuspenseQuery(userOptions);
  const logout = useLogout();

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <H2>Profile</H2>
        <Button
          variant="destructive"
          onClick={() => logout.mutate()}
          disabled={logout.isPending}
        >
          {logout.isPending ? 'Logging out…' : 'Log out'}
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Account</CardTitle>
          <CardDescription>Your email address</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm font-medium">{user.email}</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Update Email</CardTitle>
          <CardDescription>Change the email address on your account</CardDescription>
        </CardHeader>
        <CardContent>
          <UpdateEmailForm />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Update Password</CardTitle>
          <CardDescription>Change your account password</CardDescription>
        </CardHeader>
        <CardContent>
          <UpdatePasswordForm />
        </CardContent>
      </Card>

      <Card className="border-destructive">
        <CardHeader>
          <CardTitle>Danger Zone</CardTitle>
          <CardDescription>Permanently delete your account and all data</CardDescription>
        </CardHeader>
        <CardContent>
          <DeleteAccountForm />
        </CardContent>
      </Card>
    </div>
  );
};

export const Route = createFileRoute('/profile')({
  component: ProfilePage,
  loader: async ({ context }) => {
    await context.queryClient.ensureQueryData(userOptions);
  },
});