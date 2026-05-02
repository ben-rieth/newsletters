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

interface VerificationCodeInputProps {
  onVerify: (code: string) => void;
  isPending: boolean;
  error: Error | null | undefined;
  onResend: (onSuccess: () => void) => void;
  isResending: boolean;
}

export const VerificationCodeInput = ({
  onVerify,
  isPending,
  error,
  onResend,
  isResending,
}: VerificationCodeInputProps) => {
  const [code, setCode] = useState('');
  const { remaining, isRunning, start: startCountdown } = useCountdown(30);

  const handleChange = (value: string) => {
    setCode(value);
    if (value.length === 8) {
      onVerify(value);
    }
  };

  const handleResend = () => {
    onResend(() => {
      setCode('');
      startCountdown();
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-center">
        <InputOTP
          maxLength={8}
          value={code}
          onChange={handleChange}
          disabled={isPending}
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
