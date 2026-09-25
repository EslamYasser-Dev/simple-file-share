import type { Metadata } from "next";
import Link from "next/link";
import {
  Callout,
  CodeBlock,
  DocsShell,
  P,
  ProseH2,
  UL,
} from "../../../../components/site";
import { APP_URL, GITHUB_URL } from "../../../../lib/site";

export const metadata: Metadata = {
  title: "البدء",
  description:
    "ثبّت FileShare محليًا، أو عبر Docker، أو انشره — ثم ابدأ خطواتك الأولى في التطبيق.",
};

const guides = [
  {
    href: "/ar/getting-started/local/",
    title: "التطوير المحلي",
    body: "شغّل واجهة Go البرمجية والواجهة الرسومية على جهازك مع إعادة التحميل الفوري.",
    tag: "المطورون",
  },
  {
    href: "/ar/getting-started/docker/",
    title: "Docker",
    body: "أمر واحد مع Compose، أو ابنِ وشغّل حاوية واحدة.",
    tag: "التشغيل",
  },
  {
    href: "/ar/getting-started/first-steps/",
    title: "الخطوات الأولى",
    body: "سجّل الدخول، وارفع الملفات، وشارك الروابط، وادعُ فريقك.",
    tag: "المنتج",
  },
  {
    href: "/ar/getting-started/deploy/",
    title: "النشر",
    body: "مضيف Docker شامل، أو تقسيم بين واجهة ثابتة وواجهة API منفصلة.",
    tag: "الإنتاج",
  },
] as const;

export default function GettingStartedPageAr() {
  return (
    <DocsShell
      locale="ar"
      eyebrow="البدء"
      title="أطلق FileShare في دقائق"
      description="اختر مسارك: تطوير محلي، أو تشغيل Docker، أو الانتقال مباشرة إلى الإنتاج. كل دليل قصير وجاهز للنسخ واللصق."
      aside={
        <div className="panel p-5">
          <p className="eyebrow">المتطلبات السابقة</p>
          <ul className="mt-3 space-y-2 text-sm text-muted">
            <li>Go 1.25+</li>
            <li>Node.js 20+</li>
            <li>Docker (اختياري)</li>
            <li>Git</li>
          </ul>
          <a
            href={`${GITHUB_URL}#readme`}
            target="_blank"
            rel="noreferrer"
            className="mt-5 inline-flex text-sm text-accent transition hover:text-paper"
          >
            README الكامل ←
          </a>
        </div>
      }
    >
      <div className="grid gap-4 sm:grid-cols-2">
        {guides.map((g) => (
          <Link
            key={g.href}
            href={g.href}
            className="panel group p-6 transition hover:border-accent/40"
          >
            <span className="font-mono text-[10px] uppercase tracking-wider text-muted">
              {g.tag}
            </span>
            <h2 className="mt-3 font-display text-lg font-semibold text-paper">
              {g.title}
            </h2>
            <p className="mt-2 text-sm leading-relaxed text-muted">{g.body}</p>
            <span className="mt-4 inline-block text-sm text-accent transition group-hover:text-paper">
              فتح الدليل ←
            </span>
          </Link>
        ))}
      </div>

      <ProseH2>المسار السريع</ProseH2>
      <P>
        وقتك ضيق؟ استخدم Docker Compose من جذر المستودع — يبني الواجهة
        الرسومية والواجهة البرمجية معًا ويعرّض HTTP على المنفذ{" "}
        <code>22010</code>.
      </P>
      <CodeBlock
        label="الطرفية"
        code={`git clone ${GITHUB_URL}.git
cd simple-file-share
export JWT_SECRET="$(openssl rand -hex 32)"
export ADMIN_PASSWORD='your-strong-password'
docker compose up --build
# open http://localhost:22010`}
      />
      <Callout>
        يتطلب Compose تعيين <code>JWT_SECRET</code> (عشوائي، 32 حرفًا فأكثر)
        و<code>ADMIN_PASSWORD</code> — لا يوجد تسجيل دخول افتراضي، والتسجيل
        معطّل ما لم تضبط <code>ENABLE_SIGNUP=true</code>. راجع{" "}
        <Link href="/ar/getting-started/docker/">دليل Docker</Link>.
      </Callout>

      <ProseH2>ما تحصل عليه</ProseH2>
      <UL
        items={[
          "ملف Go تنفيذي واحد (أو حاوية) يقدّم الواجهة البرمجية وواجهة React",
          "رفع قابل للاستئناف مع الإيقاف والمتابعة واسترداد المهام المعلّقة",
          "نقل مباشر في الشبكة المحلية عبر إشارات WebRTC",
          "أدوار وحصص وروابط مشاركة ووحدة إدارة",
          "واجهة إنجليزية/عربية مع دعم كامل لـ RTL",
          "قرص محلي أو تخزين متوافق مع S3",
        ]}
      />

      <ProseH2>بعد التثبيت</ProseH2>
      <UL
        items={[
          <>
            اتبع <Link href="/ar/getting-started/first-steps/">الخطوات الأولى</Link>{" "}
            للرفع والمشاركة ودعوة المستخدمين.
          </>,
          <>
            وجّه تطبيق الجوال بمتغير <code>EXPO_PUBLIC_API_URL</code> (انظر
            README الرئيسي).
          </>,
          <>
            بعد تشغيل النظام، افتح التطبيق على <a href={APP_URL}>{APP_URL}</a>.
          </>,
        ]}
      />
    </DocsShell>
  );
}
