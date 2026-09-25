import type { Metadata } from "next";
import Link from "next/link";
import {
  Callout,
  CodeBlock,
  DocsShell,
  P,
  ProseH2,
  Step,
  UL,
} from "../../../../../components/site";

export const metadata: Metadata = {
  title: "الخطوات الأولى",
  description:
    "سجّل الدخول، وارفع الملفات، وأنشئ مشاركات، وأدر المستخدمين، واستخدم مشاركة الشبكة.",
};

export default function FirstStepsPageAr() {
  return (
    <DocsShell
      locale="ar"
      eyebrow="البدء / الخطوات الأولى"
      title="الخطوات الأولى في التطبيق"
      description="بعد تشغيل النظام — محليًا أو عبر Docker — هذه هي سير العمل الأساسية."
      aside={
        <div className="panel p-5">
          <p className="eyebrow">في هذه الصفحة</p>
          <ul className="mt-3 space-y-2 text-sm text-muted">
            <li>
              <a href="#sign-in">تسجيل الدخول</a>
            </li>
            <li>
              <a href="#upload">الرفع والاستئناف</a>
            </li>
            <li>
              <a href="#share">روابط المشاركة</a>
            </li>
            <li>
              <a href="#team">الفريق والأدوار</a>
            </li>
            <li>
              <a href="#lan">مشاركة الشبكة</a>
            </li>
            <li>
              <a href="#language">اللغة</a>
            </li>
          </ul>
          <div className="mt-5 space-y-2 text-sm">
            <Link
              href="/ar/getting-started/"
              className="block text-accent transition hover:text-paper"
            >
              → كل الأدلة
            </Link>
            <Link
              href="/ar/getting-started/deploy/"
              className="block text-accent transition hover:text-paper"
            >
              التالي: النشر ←
            </Link>
          </div>
        </div>
      }
    >
      <section id="sign-in">
        <ProseH2>1. تسجيل الدخول</ProseH2>
        <UL
          items={[
            <>
              <strong>المصادقة معطّلة (تطوير):</strong> تدخل مباشرة على واجهة
              المدير النظام — دون نموذج تسجيل دخول.
            </>,
            <>
              <strong>المصادقة مفعّلة:</strong> استخدم بيانات{" "}
              <code>ADMIN_USERNAME</code>/<code>ADMIN_PASSWORD</code> الأولية،
              أو سجّل حسابًا جديدًا إن كان <code>ENABLE_SIGNUP=true</code>{" "}
              (أول حساب يصبح مديرًا).
            </>,
          ]}
        />
        <Callout>
          التسجيل الذاتي يتحكّم فيه <code>ENABLE_SIGNUP</code>. عطّله بعد
          دعوة فريقك إن أردت المستخدمين بالدعوة فقط.
        </Callout>
      </section>

      <section id="upload">
        <ProseH2>2. الرفع والاستئناف</ProseH2>
        <Step locale="ar" n="1" title="الرفع">
          <P>
            استخدم <strong>رفع</strong> في الشريط الجانبي أو أفلت الملفات فوق
            قائمة الملفات. يعرض التقدّم السرعة والنسبة؛ ودفعات الملفات المتعدّة
            تُوحَّد في مؤشر واحد.
          </P>
        </Step>
        <Step locale="ar" n="2" title="الإيقاف أو الإلغاء">
          <P>
            أثناء الرفع، اضغط <strong>إيقاف</strong> في الإشعار. تتوقّف المقاطع
            بين الإرسالات؛ و<strong>المتابعة</strong> تستأنف من الموضع المحفوظ.
            الإلغاء يقطع الجلسة ويمسحها.
          </P>
        </Step>
        <Step locale="ar" n="3" title="الاسترداد بعد إعادة التحميل">
          <P>
            تظهر المهام المقاطعة تحت <strong>الرفع المعلّق</strong>. تعيد
            المتابعة فتح منتقي الملفات وتحقّق من المسار والاسم والحجم قبل
            الاستمرار.
          </P>
          <Callout>
            الملفات فوق عتبة الرفع المقسّم (1&nbsp;MB) تستخدم جلسات مقسّمة.
            الملفات الأصغر تُرفع بطلب واحد.
          </Callout>
        </Step>
      </section>

      <section id="share">
        <ProseH2>3. روابط المشاركة</ProseH2>
        <UL
          items={[
            "افتح ملفًا أو مجلدًا ← مشاركة",
            "انسخ الرابط (عنوان يعتمد على رمز عشوائي)",
            "ألغِ المشاركة عند الانتهاء",
            "المجلد المشترك: كل مستخدم مسجّل يستطيع القراءة؛ والكتابة للمديرين فقط",
          ]}
        />
        <P>المجلدات تُنزَّل كملف ZIP أثناء الطفرة — دون أرشيف مؤقت على القرص.</P>
      </section>

      <section id="team">
        <ProseH2>4. الفريق والأدوار</ProseH2>
        <UL
          items={[
            <>
              <strong>وحدة الإدارة</strong> — قائمة المستخدمين والاستخدام والحصص
            </>,
            <>
              الأدوار: <code>admin</code> و<code>member</code>، إضافة إلى أدوار
              مخصّصة بمصفوفة صلاحيات
            </>,
            "حصص تخزين لكل مستخدم وإعادة تعيين كلمات المرور",
            "مجلدات منزلية خاصة؛ مساحات المستخدمين الآخرين تعيد 403",
          ]}
        />
        <P>مسار الدعوة المعتاد:</P>
        <CodeBlock
          code={`1. Enable ENABLE_SIGNUP (or create users as admin)
2. Share the app URL
3. New users register → land in their private home
4. Assign roles / quotas from the admin console`}
        />
      </section>

      <section id="lan">
        <ProseH2>5. مشاركة الشبكة (من جهاز لآخر)</ProseH2>
        <UL
          items={[
            "افتح مشاركة الشبكة في الشريط الجانبي",
            "يجب أن يصل الجهازان إلى الخادم نفسه (نفس الشبكة المحلية / VPN)",
            "اختر جهازًا ← وأرسل ملفًا؛ البايتات تنتقل مباشرة عبر WebRTC",
            "الخادم يوسط فقط في إشارات SDP/ICE — لا في محتوى الملفات",
          ]}
        />
        <Callout>
          يمكنك إيقاف مشاركة الشبكة في أي وقت من{" "}
          <strong>الإعدادات ← مشاركة الشبكة</strong>. يختفي عنصر التنقّل وتُغلق
          بث الإشارات.
        </Callout>
      </section>

      <section id="language">
        <ProseH2>6. اللغة والمظهر</ProseH2>
        <UL
          items={[
            "بدّل بين الإنجليزية والعربية من الترويسة (دعم كامل لـ RTL)",
            "تُحفظ تفضيلاتك في المتصفح",
            "الإعدادات: مظهر فاتح/داكن/تلقائي وعدد عمليات الرفع المتوازية",
          ]}
        />
      </section>

      <ProseH2>استكشف بعد ذلك</ProseH2>
      <UL
        items={[
          <>
            <Link href="/ar/getting-started/deploy/">النشر إلى الإنتاج</Link>
          </>,
          <>
            مرجع الواجهة البرمجية في README الخاص بالمستودع و<code>/swagger</code>{" "}
            على خادم يعمل
          </>,
          <>
            <Link href="/ar/getting-started/local/">التطوير المحلي</Link>{" "}
            للمشاركة في تطوير المشروع
          </>,
        ]}
      />
    </DocsShell>
  );
}
