import { useForm } from '@tanstack/react-form';
import { z } from 'zod';
import { Button } from '#/components/ui/button';
import { FieldError } from '#/components/ui/field';
import { FormField } from '#/components/ui/form-field';
import { Input } from '#/components/ui/input';
import { NumberField } from '#/components/ui/number-field';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import { SettingsRow, SettingsSection } from '#/components/SettingsRow';
import { DAY_NAMES } from '../lib/format';
import type { Frequency } from '../queries/newsletters';
import { TimeField } from './TimeField';
import { TimezoneSelect } from './TimezoneSelect';

const VALID_TIMEZONES = new Set(Intl.supportedValuesOf('timeZone'));

export const newsletterFormSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  frequency: z.enum(['daily', 'weekly', 'monthly']),
  sendHour: z.number().min(0, 'Must be 0–23').max(23, 'Must be 0–23'),
  sendMinute: z.number().min(0, 'Must be 0–59').max(59, 'Must be 0–59'),
  sendDay: z.number().int().nonnegative().optional(),
  sendTimezone: z.string().refine((tz) => VALID_TIMEZONES.has(tz), {
    message: 'Invalid timezone',
  }),
});

export type NewsletterFormValues = z.infer<typeof newsletterFormSchema>;

const FREQUENCY_OPTIONS = [
  { label: 'Daily', value: 'daily' },
  { label: 'Weekly', value: 'weekly' },
  { label: 'Monthly', value: 'monthly' },
];

type Props = {
  defaultValues: NewsletterFormValues;
  onSubmit: (values: NewsletterFormValues) => Promise<void>;
  isPending?: boolean;
  submitLabel?: string;
  error?: string;
  layout?: 'stacked' | 'settings';
};

export const NewsletterForm = ({
  defaultValues,
  onSubmit,
  isPending = false,
  submitLabel = 'Save',
  error,
  layout = 'stacked',
}: Props) => {
  const form = useForm({
    defaultValues,
    validators: { onChange: newsletterFormSchema },
    onSubmit: async ({ value }) => onSubmit(value),
  });

  const fields = (
    <>
      <form.Field name="name">
        {(field) => (
          <FormField field={field} label="Name">
            <Input
              id={field.name}
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              placeholder="My Newsletter"
              aria-invalid={
                field.state.meta.isBlurred && field.state.meta.errors.length > 0
              }
            />
          </FormField>
        )}
      </form.Field>
      <form.Field name="frequency">
        {(field) => (
          <FormField field={field} label="Frequency">
            <Select
              value={field.state.value}
              onValueChange={(value) => {
                field.handleChange(value as Frequency);
                form.setFieldValue('sendDay', undefined);
              }}
              items={FREQUENCY_OPTIONS}
            >
              <SelectTrigger id={field.name} className="w-full">
                <SelectValue placeholder="Select frequency" />
              </SelectTrigger>
              <SelectContent>
                {FREQUENCY_OPTIONS.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>
        )}
      </form.Field>

      <form.Subscribe selector={(state) => state.values.frequency}>
        {(frequency) =>
          frequency !== 'daily' && (
            <form.Field name="sendDay">
              {(field) =>
                frequency === 'weekly' ? (
                  <FormField field={field} label="Send on Day">
                    <Select
                      value={
                        field.state.value !== undefined
                          ? String(field.state.value)
                          : ''
                      }
                      onValueChange={(value) =>
                        field.handleChange(Number(value))
                      }
                      items={DAY_NAMES.map((day, i) => ({
                        value: String(i),
                        label: day,
                      }))}
                    >
                      <SelectTrigger id={field.name} className="w-full">
                        <SelectValue placeholder="Select a day" />
                      </SelectTrigger>
                      <SelectContent>
                        {DAY_NAMES.map((day, i) => (
                          <SelectItem key={day} value={String(i)}>
                            {day}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </FormField>
                ) : (
                  <FormField field={field} label="Day of Month (1–31)">
                    <NumberField
                      id={field.name}
                      min={1}
                      max={31}
                      value={field.state.value ?? null}
                      onValueChange={(value) =>
                        field.handleChange(value ?? undefined)
                      }
                      aria-invalid={
                        field.state.meta.isBlurred &&
                        field.state.meta.errors.length > 0
                      }
                    />
                  </FormField>
                )
              }
            </form.Field>
          )
        }
      </form.Subscribe>

      <form.Field name="sendTimezone">
        {(field) => (
          <FormField field={field} label="Timezone">
            <TimezoneSelect
              id={field.name}
              value={field.state.value}
              onValueChange={(value) => field.handleChange(value)}
            />
          </FormField>
        )}
      </form.Field>

      <form.Subscribe
        selector={(state) => ({
          hour: state.values.sendHour,
          minute: state.values.sendMinute,
        })}
      >
        {({ hour, minute }) => (
          <TimeField
            id="sendTime"
            label="Send Time"
            hour={hour}
            minute={minute}
            onChange={(h, m) => {
              form.setFieldValue('sendHour', h);
              form.setFieldValue('sendMinute', m);
            }}
          />
        )}
      </form.Subscribe>
    </>
  );

  const submitBlock = (
    <div className="space-y-3">
      {error && <FieldError>{error}</FieldError>}
      <Button type="submit" disabled={isPending}>
        {isPending ? `${submitLabel}…` : submitLabel}
      </Button>
    </div>
  );

  return (
    <form
      className="space-y-5"
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      {layout === 'settings' ? (
        <SettingsSection>
          <SettingsRow
            title="Newsletter"
            description="Newsletter name and sending schedule."
          >
            <div className="space-y-5">
              {fields}
              {submitBlock}
            </div>
          </SettingsRow>
        </SettingsSection>
      ) : (
        <>
          {fields}
          {submitBlock}
        </>
      )}
    </form>
  );
};
