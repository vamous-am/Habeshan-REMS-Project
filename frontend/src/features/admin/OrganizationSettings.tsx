import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "../../components/Button";
import { Card } from "../../components/Card";
import { ErrorState } from "../../components/ErrorState";
import { Input } from "../../components/Input";
import { StatusBadge } from "../../components/StatusBadge";
import { fetchOrg, updateOrg, type OrgDTO } from "../../lib/api/adminClient";
import { extractAuthErrorMessage } from "../../lib/api/authClient";
import { useToast } from "../../components/useToast";

const orgSchema = z.object({
  name: z.string().min(2, "Organization name must be at least 2 characters"),
  currency: z.string().min(3, "Enter a currency code (e.g. ETB)"),
  timezone: z.string().min(3, "Enter a timezone (e.g. Africa/Addis_Ababa)"),
});

type OrgForm = z.infer<typeof orgSchema>;

function subscriptionVariant(
  status: OrgDTO["subscription_status"]
): "verified" | "pending" | "rejected" {
  if (status === "active") return "verified";
  if (status === "trial") return "pending";
  return "rejected";
}

export default function OrganizationSettings() {
  const { showSuccess } = useToast();
  const [org, setOrg] = useState<OrgDTO | null>(null);
  const [loading, setLoading] = useState(true);
  const [apiError, setApiError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const form = useForm<OrgForm>({
    resolver: zodResolver(orgSchema),
    defaultValues: { name: "", currency: "", timezone: "" },
  });

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      try {
        const data = await fetchOrg();
        if (!cancelled) {
          setOrg(data);
          form.reset({
            name: data.name,
            currency: data.currency,
            timezone: data.timezone,
          });
        }
      } catch (err) {
        if (!cancelled) {
          setApiError(extractAuthErrorMessage(err));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [form]);

  async function handleSubmit(values: OrgForm) {
    setSaving(true);
    setApiError(null);
    try {
      const updated = await updateOrg({
        name: values.name.trim(),
        currency: values.currency.trim(),
        timezone: values.timezone.trim(),
      });
      setOrg(updated);
      showSuccess("Organization settings saved.");
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setSaving(false);
    }
  }

  return (
    <section>
      <div className="mb-6">
        <h2 className="font-display text-xl font-semibold text-ink">
          Organization settings
        </h2>
        <p className="mt-1 text-sm text-ink500">
          Update your organization profile and locale defaults.
        </p>
      </div>

      {apiError && (
        <div className="mb-4">
          <ErrorState message={apiError} />
        </div>
      )}

      {loading ? (
        <p className="text-sm text-ink500">Loading organization…</p>
      ) : (
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_280px]">
          <Card title="General" className="max-w-xl">
            <form
              className="flex flex-col gap-4"
              onSubmit={form.handleSubmit(handleSubmit)}
              noValidate
            >
              <Input
                label="Organization name"
                error={form.formState.errors.name?.message}
                {...form.register("name")}
              />
              <Input
                label="Currency"
                hint="ISO code, e.g. ETB"
                error={form.formState.errors.currency?.message}
                {...form.register("currency")}
              />
              <Input
                label="Timezone"
                hint="IANA timezone, e.g. Africa/Addis_Ababa"
                error={form.formState.errors.timezone?.message}
                {...form.register("timezone")}
              />
              <Button type="submit" loading={saving}>
                Save changes
              </Button>
            </form>
          </Card>

          {org && (
            <aside className="rounded border border-ink/10 bg-paper-dim p-6 shadow-card">
              <h3 className="font-display text-sm font-semibold uppercase tracking-wide text-ink500">
                Subscription
              </h3>
              <dl className="mt-4 space-y-4 text-sm">
                <div>
                  <dt className="text-ink500">Status</dt>
                  <dd className="mt-1">
                    <StatusBadge
                      label={org.subscription_status}
                      variant={subscriptionVariant(org.subscription_status)}
                    />
                  </dd>
                </div>
                <div>
                  <dt className="text-ink500">Seat count</dt>
                  <dd className="mt-1 font-medium text-ink900">
                    {org.seat_count}
                  </dd>
                </div>
                <div>
                  <dt className="text-ink500">Organization ID</dt>
                  <dd className="mt-1 font-mono text-xs text-ink500 break-all">
                    {org.id}
                  </dd>
                </div>
              </dl>
            </aside>
          )}
        </div>
      )}
    </section>
  );
}
