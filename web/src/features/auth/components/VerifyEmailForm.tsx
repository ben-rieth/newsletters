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
