import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { FieldError } from '#/components/ui/field';
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSeparator,
  InputOTPSlot,
} from '#/components/ui/input-otp';
import { getErrorMessage } from '#/lib/errors';
import useCountdown from '../lib/useCountdown';
import useResendVerification from '../queries/hooks/useResendVerification';
import useVerifyEmail from '../queries/hooks/useVerifyEmail';

export const VerifyEmailForm = () => {
  const [code, setCode] = useState('');

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
  const { remaining, isRunning, start: startCountdown } = useCountdown(30);

  const handleChange = (value: string) => {
    setCode(value);
    if (value.length === 8) {
      verify(value);
    }
  };

  const handleResend = () => {
    resend(undefined, {
      onSuccess: () => {
        setCode('');
        startCountdown();
      },
    });
  };

  const error = verifyError ?? resendError;

  return (
    <div className="space-y-4">
      <div className="flex justify-center">
        <InputOTP
          maxLength={8}
          value={code}
          onChange={handleChange}
          disabled={isVerifying}
        >
          <InputOTPGroup>
            <InputOTPSlot index={0} />
            <InputOTPSlot index={1} />
            <InputOTPSlot index={2} />
            <InputOTPSlot index={3} />
          </InputOTPGroup>
          <InputOTPSeparator />
          <InputOTPGroup>
            <InputOTPSlot index={4} />
            <InputOTPSlot index={5} />
            <InputOTPSlot index={6} />
            <InputOTPSlot index={7} />
          </InputOTPGroup>
        </InputOTP>
      </div>

      {error && <FieldError>{getErrorMessage(error)}</FieldError>}

      <div className="text-center text-sm text-muted-foreground">
        {isRunning ? (
          <span>Resend code in {remaining}s</span>
        ) : (
          <Button
            variant="link"
            className="h-auto p-0 text-sm"
            onClick={handleResend}
            disabled={isResending}
          >
            {isResending ? 'Sending…' : 'Resend code'}
          </Button>
        )}
      </div>
    </div>
  );
};
