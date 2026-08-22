import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { z } from "zod";
import { Button } from "../../components/Button";
import { Card } from "../../components/Card";
import { ErrorState } from "../../components/ErrorState";
import { Input } from "../../components/Input";
import {
  extractAuthErrorMessage,
  login,
  lookupOrgs,
  persistAuthSession,
  type OrgSummary,
} from "../../lib/api/authClient";

const emailSchema = z.object({
  email: z.string().email("Enter a valid email address"),
});

const passwordSchema = z.object({
  password: z.string().min(1, "Password is required"),
});

type EmailForm = z.infer<typeof emailSchema>;
type PasswordForm = z.infer<typeof passwordSchema>;

type LoginStep = "email" | "org" | "password";

export default function Login() {
  const navigate = useNavigate();
  const [step, setStep] = useState<LoginStep>("email");
  const [email, setEmail] = useState("");
  const [orgs, setOrgs] = useState<OrgSummary[]>([]);
  const [selectedOrgId, setSelectedOrgId] = useState("");
  const [apiError, setApiError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const emailForm = useForm<EmailForm>({
    resolver: zodResolver(emailSchema),
    defaultValues: { email: "" },
  });

  const passwordForm = useForm<PasswordForm>({
    resolver: zodResolver(passwordSchema),
    defaultValues: { password: "" },
  });

  async function handleEmailSubmit(values: EmailForm) {
    setLoading(true);
    setApiError(null);
    try {
      const { orgs: foundOrgs } = await lookupOrgs(values.email);
      setEmail(values.email);

      if (foundOrgs.length === 0) {
        setApiError("No account found for this email.");
        return;
      }

      if (foundOrgs.length === 1) {
        setSelectedOrgId(foundOrgs[0].org_id);
        setOrgs(foundOrgs);
        setStep("password");
        return;
      }

      setOrgs(foundOrgs);
      setSelectedOrgId(foundOrgs[0].org_id);
      setStep("org");
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  async function handlePasswordSubmit(values: PasswordForm) {
    setLoading(true);
    setApiError(null);
    try {
      const session = await login({
        email,
        password: values.password,
        org_id: selectedOrgId,
      });
      persistAuthSession(session);
      navigate("/tasks");
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  function handleOrgContinue() {
    if (!selectedOrgId) {
      setApiError("Select an organization to continue.");
      return;
    }
    setApiError(null);
    setStep("password");
  }

  function handleBack() {
    setApiError(null);
    passwordForm.reset();
    if (step === "password" && orgs.length > 1) {
      setStep("org");
      return;
    }
    setStep("email");
    setOrgs([]);
    setSelectedOrgId("");
  }

  const selectedOrg = orgs.find((org) => org.org_id === selectedOrgId);

  return (
    <div className="flex min-h-screen items-center justify-center bg-paper p-4">
      <Card
        title="Sign in"
        subtitle={
          step === "email"
            ? "Enter your email to continue"
            : step === "org"
              ? "Choose your organization"
              : `Signing in as ${email}${selectedOrg ? ` · ${selectedOrg.org_name}` : ""}`
        }
      >
        {apiError && (
          <div className="mb-4">
            <ErrorState message={apiError} />
          </div>
        )}

        {step === "email" && (
          <form
            className="flex flex-col gap-4"
            onSubmit={emailForm.handleSubmit(handleEmailSubmit)}
            noValidate
          >
            <Input
              label="Email"
              type="email"
              autoComplete="email"
              error={emailForm.formState.errors.email?.message}
              {...emailForm.register("email")}
            />
            <Button type="submit" loading={loading} className="w-full">
              Continue
            </Button>
          </form>
        )}

        {step === "org" && (
          <div className="flex flex-col gap-4">
            <fieldset className="flex flex-col gap-2">
              <legend className="mb-1 text-sm font-medium text-ink">
                Organization
              </legend>
              {orgs.map((org) => (
                <label
                  key={org.org_id}
                  className={[
                    "flex cursor-pointer items-center gap-3 rounded border px-3 py-2",
                    selectedOrgId === org.org_id
                      ? "border-ink bg-paper-dim"
                      : "border-ink/20",
                  ].join(" ")}
                >
                  <input
                    type="radio"
                    name="org"
                    value={org.org_id}
                    checked={selectedOrgId === org.org_id}
                    onChange={() => setSelectedOrgId(org.org_id)}
                    className="accent-ink"
                  />
                  <span className="text-sm text-ink">{org.org_name}</span>
                </label>
              ))}
            </fieldset>
            <div className="flex gap-3">
              <Button
                type="button"
                variant="secondary"
                onClick={handleBack}
                className="flex-1"
              >
                Back
              </Button>
              <Button
                type="button"
                onClick={handleOrgContinue}
                className="flex-1"
              >
                Continue
              </Button>
            </div>
          </div>
        )}

        {step === "password" && (
          <form
            className="flex flex-col gap-4"
            onSubmit={passwordForm.handleSubmit(handlePasswordSubmit)}
            noValidate
          >
            <Input
              label="Password"
              type="password"
              autoComplete="current-password"
              error={passwordForm.formState.errors.password?.message}
              {...passwordForm.register("password")}
            />
            <div className="flex gap-3">
              <Button
                type="button"
                variant="secondary"
                onClick={handleBack}
                className="flex-1"
              >
                Back
              </Button>
              <Button type="submit" loading={loading} className="flex-1">
                Sign in
              </Button>
            </div>
          </form>
        )}

        <p className="mt-6 text-center text-sm text-ink500">
          No account?{" "}
          <Link to="/register" className="text-ink underline underline-offset-2">
            Register
          </Link>
          {" · "}
          <Link
            to="/forgot-password"
            className="text-ink underline underline-offset-2"
          >
            Forgot password
          </Link>
        </p>
      </Card>
    </div>
  );
}
