import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { z } from "zod";
import {
  extractAuthErrorMessage,
  login,
  lookupOrgs,
  persistAuthSession,
  type OrgSummary,
} from "../../lib/api/authClient";
import { AuthField, AuthFieldGroup } from "./AuthField";
import { AuthButton, AuthError, AuthLayout } from "./AuthLayout";

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
    <AuthLayout
      title="Welcome back"
      subtitle={
        step === "email"
          ? "Sign in to your organization workspace. Enter your email to continue."
          : step === "org"
            ? "Choose your organization"
            : `Signing in as ${email}${selectedOrg ? ` · ${selectedOrg.org_name}` : ""}`
      }
      footer={
        <>
          <p>
            No account?{" "}
            <Link
              to="/register"
              className="font-semibold text-ochre-light underline-offset-2 hover:underline"
            >
              Register
            </Link>
          </p>
          <p className="mt-2">
            <Link
              to="/forgot-password"
              className="text-ochre-light underline-offset-2 hover:underline"
            >
              Forgot password
            </Link>
          </p>
        </>
      }
    >
      {apiError && <AuthError message={apiError} />}

      {step === "email" && (
        <form
          className="flex flex-col gap-5"
          onSubmit={emailForm.handleSubmit(handleEmailSubmit)}
          noValidate
        >
          <AuthFieldGroup>
            <AuthField
              label="Email"
              type="email"
              autoComplete="email"
              placeholder="you@agency.et"
              error={emailForm.formState.errors.email?.message}
              {...emailForm.register("email")}
            />
          </AuthFieldGroup>
          <AuthButton type="submit" loading={loading} className="w-full">
            Continue
          </AuthButton>
        </form>
      )}

      {step === "org" && (
        <div className="flex flex-col gap-5">
          <fieldset>
            <legend className="sr-only">Organization</legend>
            <AuthFieldGroup>
              {orgs.map((org) => (
                <label
                  key={org.org_id}
                  className={[
                    "flex cursor-pointer items-center gap-3 px-4 py-3 sm:px-5",
                    selectedOrgId === org.org_id
                      ? "bg-ochre/10"
                      : "hover:bg-white/[0.03]",
                  ].join(" ")}
                >
                  <input
                    type="radio"
                    name="org"
                    value={org.org_id}
                    checked={selectedOrgId === org.org_id}
                    onChange={() => setSelectedOrgId(org.org_id)}
                    className="accent-ochre"
                  />
                  <span
                    className={[
                      "text-base",
                      selectedOrgId === org.org_id
                        ? "font-medium text-paper"
                        : "text-auth-muted",
                    ].join(" ")}
                  >
                    {org.org_name}
                  </span>
                </label>
              ))}
            </AuthFieldGroup>
          </fieldset>
          <div className="flex gap-3">
            <AuthButton
              type="button"
              variant="secondary"
              onClick={handleBack}
              className="flex-1"
            >
              Back
            </AuthButton>
            <AuthButton
              type="button"
              onClick={handleOrgContinue}
              className="flex-1"
            >
              Continue
            </AuthButton>
          </div>
        </div>
      )}

      {step === "password" && (
        <form
          className="flex flex-col gap-5"
          onSubmit={passwordForm.handleSubmit(handlePasswordSubmit)}
          noValidate
        >
          <AuthFieldGroup>
            <AuthField
              label="Password"
              type="password"
              autoComplete="current-password"
              placeholder="Password"
              passwordToggle
              error={passwordForm.formState.errors.password?.message}
              {...passwordForm.register("password")}
            />
          </AuthFieldGroup>
          <div className="flex gap-3">
            <AuthButton
              type="button"
              variant="secondary"
              onClick={handleBack}
              className="flex-1"
            >
              Back
            </AuthButton>
            <AuthButton type="submit" loading={loading} className="flex-1">
              Sign in
            </AuthButton>
          </div>
        </form>
      )}
    </AuthLayout>
  );
}
