import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { Button } from '#/components/ui/button'
import { Field, FieldError, FieldLabel } from '#/components/ui/field'
import { FormField } from '#/components/ui/form-field'
import { Input } from '#/components/ui/input'
import { NumberField } from '#/components/ui/number-field'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import { DAY_NAMES } from '../lib/format'
import type { Frequency } from '../queries/newsletters'
import { TimezoneSelect } from './TimezoneSelect'

const VALID_TIMEZONES = new Set(Intl.supportedValuesOf('timeZone'))

export const newsletterFormSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  frequency: z.enum(['daily', 'weekly', 'monthly']),
  sendHour: z.number().min(0, 'Must be 0–23').max(23, 'Must be 0–23'),
  sendMinute: z.number().min(0, 'Must be 0–59').max(59, 'Must be 0–59'),
  sendDay: z.number().int().nonnegative().optional(),
  sendTimezone: z.string().refine((tz) => VALID_TIMEZONES.has(tz), {
    message: 'Invalid timezone',
  }),
})

export type NewsletterFormValues = z.infer<typeof newsletterFormSchema>

const FREQUENCY_OPTIONS = [
  { label: 'Daily', value: 'daily' },
  { label: 'Weekly', value: 'weekly' },
  { label: 'Monthly', value: 'monthly' },
]

type Props = {
  defaultValues: NewsletterFormValues
  onSubmit: (values: NewsletterFormValues) => Promise<void>
  isPending?: boolean
  submitLabel?: string
  error?: string
  renderActions?: (isPending: boolean) => React.ReactNode
}

export const NewsletterForm = ({
  defaultValues,
  onSubmit,
  isPending = false,
  submitLabel = 'Save',
  error,
}: Props) => {
  const form = useForm({
    defaultValues,
    validators: { onChange: newsletterFormSchema },
    onSubmit: async ({ value }) => onSubmit(value),
  })

  return (
    <form
      className="space-y-3"
      onSubmit={(e) => {
        e.preventDefault()
        form.handleSubmit()
      }}
    >
      {/* Name */}
      <form.Field name="name">
        {(field) => (
          <FormField field={field} label="Name">
            <Input
              id={field.name}
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              placeholder="My Newsletter"
              aria-invalid={field.state.meta.isBlurred && field.state.meta.errors.length > 0}
            />
          </FormField>
        )}
      </form.Field>

      {/* Frequency */}
      <form.Field name="frequency">
        {(field) => (
          <FormField field={field} label="Frequency">
            <Select
              value={field.state.value}
              onValueChange={(value) => {
                field.handleChange(value as Frequency)
                form.setFieldValue('sendDay', undefined)
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

      {/* Send day (conditional on frequency) */}
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
                      aria-invalid={field.state.meta.isBlurred && field.state.meta.errors.length > 0}
                    />
                  </FormField>
                )
              }
            </form.Field>
          )
        }
      </form.Subscribe>

      {/* Timezone */}
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

      {/* Send time */}
      <form.Subscribe selector={(state) => ({ hour: state.values.sendHour, minute: state.values.sendMinute })}>
        {({ hour, minute }) => {
          const timeValue = `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`;
          return (
            <Field>
              <FieldLabel htmlFor="sendTime">Send Time</FieldLabel>
              <Input
                id="sendTime"
                type="time"
                value={timeValue}
                onChange={(e) => {
                  const [h, m] = (e.target.value || '00:00').split(':').map(Number);
                  form.setFieldValue('sendHour', isNaN(h) ? 0 : h);
                  form.setFieldValue('sendMinute', isNaN(m) ? 0 : m);
                }}
              />
            </Field>
          );
        }}
      </form.Subscribe>

      {error && <FieldError>{error}</FieldError>}

      <Button type="submit" disabled={isPending}>
        {isPending ? `${submitLabel}…` : submitLabel}
      </Button>
    </form>
  )
}