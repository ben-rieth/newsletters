import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { Button } from '#/components/ui/button';
import { SettingsRow, SettingsSection } from '#/components/SettingsRow';
import { UpdateEmailForm } from '#/features/auth/components/UpdateEmailForm';
import { UpdatePasswordForm } from '#/features/auth/components/UpdatePasswordForm';
import { DeleteAccountForm } from '#/features/auth/components/DeleteAccountForm';
import { userOptions } from '#/features/auth/queries/user';
import useLogout from '#/features/auth/queries/hooks/useLogout';
import useExportNewsletters from '#/features/newsletters/queries/hooks/useExportNewsletters';

const ProfilePage = () => {
  const { data: user } = useSuspenseQuery(userOptions);
  const logout = useLogout();
  const exportAll = useExportNewsletters();

  return (
    <div className="mx-auto max-w-3xl space-y-10 px-6 py-8 lg:px-10">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">Profile</h1>
        <p className="text-sm text-muted-foreground">
          Manage your account details and security.
        </p>
      </header>

      <SettingsSection>
        <SettingsRow
          title="Email address"
          description="The email address associated with your account."
        >
          <p className="text-sm font-medium">{user.email}</p>
        </SettingsRow>

        <SettingsRow
          title="Update email"
          description="Change the email address on your account. You'll confirm the new address with a code."
        >
          <UpdateEmailForm />
        </SettingsRow>

        <SettingsRow
          title="Update password"
          description="Choose a new password for your account."
        >
          <UpdatePasswordForm />
        </SettingsRow>
      </SettingsSection>

      <SettingsSection>
        <SettingsRow
          title="Export newsletters"
          description="Download all of your newsletters and their feeds as a JSON file."
        >
          <Button
            variant="outline"
            onClick={() => exportAll.mutate()}
            disabled={exportAll.isPending}
          >
            {exportAll.isPending ? 'Exporting…' : 'Export all newsletters'}
          </Button>
        </SettingsRow>
      </SettingsSection>

      <SettingsSection>
        <SettingsRow
          title="Log out"
          description="Sign out of your account on this device."
        >
          <Button
            variant="outline"
            onClick={() => logout.mutate()}
            disabled={logout.isPending}
          >
            {logout.isPending ? 'Logging out…' : 'Log out'}
          </Button>
        </SettingsRow>

        <SettingsRow
          title="Delete account"
          description="Permanently delete your account and all of your data. This cannot be undone."
        >
          <DeleteAccountForm />
        </SettingsRow>
      </SettingsSection>
    </div>
  );
};

export const Route = createFileRoute('/profile')({
  component: ProfilePage,
  loader: async ({ context }) => {
    await context.queryClient.ensureQueryData(userOptions);
  },
});
