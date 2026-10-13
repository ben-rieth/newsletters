import { Link } from '@tanstack/react-router';
import { getPendingVerificationPassword } from '../lib/pendingVerificationPassword';
import useResendVerification from '../queries/hooks/useResendVerification';
import useVerifyEmail from '../queries/hooks/useVerifyEmail';
import { VerificationCodeInput } from './VerificationCodeInput';

export const VerifyEmailForm = () => {
  const {
    mutate: verify,
    isPending: isVerifying,
    error: verifyError,
  } = useVerifyEmail();
  const {
    mutate: resend,
    isPending: isResending,
    error: resendError,
  } = useResendVerification();

  if (getPendingVerificationPassword() === null) {
    return (
      <p className="text-center text-sm text-muted-foreground">
        To finish verifying your email, please{' '}
        <Link to="/sign-in" className="underline underline-offset-4">
          sign in again
        </Link>
        .
      </p>
    );
  }

  return (
    <VerificationCodeInput
      onVerify={verify}
      isPending={isVerifying}
      error={verifyError ?? resendError}
      onResend={(onSuccess) => resend(undefined, { onSuccess })}
      isResending={isResending}
    />
  );
};
