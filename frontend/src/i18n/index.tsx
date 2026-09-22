import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';

export type Lang = 'en' | 'ar';

const en = {
  // Layout / nav
  'nav.myFiles': 'My Files',
  'nav.summary': 'Summary',
  'nav.activity': 'Activity',
  'nav.upload': 'Upload',
  'nav.newFolder': 'New Folder',
  'nav.cloudStorage': 'Cloud Storage',
  'nav.language': 'Language',
  'nav.theme': 'Theme',
  'nav.config': 'Configuration',
  'nav.shared': 'Shared',
  'nav.admin': 'Users',
  'nav.signOut': 'Sign out',

  // Common
  'common.close': 'Close',
  'common.cancel': 'Cancel',
  'common.create': 'Create',
  'common.save': 'Save',
  'common.download': 'Download',
  'common.preview': 'Preview',
  'common.edit': 'Edit',
  'common.retry': 'Retry',
  'common.delete': 'Delete',
  'common.openInNewTab': 'Open in new tab',
  'common.back': 'Back',
  'common.dismiss': 'Dismiss',

  // Configuration panel
  'config.theme.system': 'System',
  'config.theme.dark': 'Dark',
  'config.theme.light': 'Light',
  'config.maxUploads': 'Max parallel uploads',
  'config.hint': 'Applies to how many files upload at the same time (1–8).',

  // Auth
  'auth.connecting': 'Connecting...',
  'auth.unableToReach': 'Unable to reach the server',
  'auth.signIn': 'Sign in',
  'auth.signInToAccess': 'Sign in to access your files',
  'auth.username': 'Username',
  'auth.password': 'Password',
  'auth.invalidCredentials': 'Invalid username or password',
  'auth.createAccount': 'Create an account',
  'auth.noAccount': "Don't have an account?",
  'auth.haveAccount': 'Already have an account?',
  'auth.signUp': 'Sign up',
  'auth.signupDisabled': 'Registration is disabled',
  'auth.confirmPassword': 'Confirm password',
  'auth.passwordsMismatch': 'Passwords do not match',
  'auth.usernameHint': '3-32 characters: letters, digits, dots, dashes, underscores',
  'auth.passwordHint': 'At least 4 characters',

  // Register
  'register.title': 'Create account',
  'register.subtitle': 'Sign up to start storing your files',
  'register.submit': 'Create account',
  'register.failed': 'Could not create the account',

  // Admin
  'admin.title': 'Users',
  'admin.subtitle': 'Manage accounts and storage usage',
  'admin.loading': 'Loading users...',
  'admin.loadFailed': 'Failed to load users',
  'admin.username': 'Username',
  'admin.role': 'Role',
  'admin.roleAdmin': 'Admin',
  'admin.roleUser': 'User',
  'admin.files': 'Files',
  'admin.size': 'Storage',
  'admin.createdAt': 'Joined',
  'admin.empty': 'No users found',

  // Shared
  'shared.readOnly': 'The shared folder is read-only',

  // New folder
  'folder.created': 'Folder created',
  'folder.nameRequired': 'Folder name is required',
  'folder.namePlaceholder': 'Folder name',

  // Home
  'home.items': '{n} item(s)',
  'home.refresh': 'Refresh',
  'home.upload': 'Upload',
  'home.uploading': 'Uploading...',
  'home.uploaded': 'Uploaded {n} file(s) successfully',
  'home.uploadCancelled': 'Upload cancelled',
  'home.cancelUpload': 'Cancel upload',
  'home.cancelUploadTitle': 'Cancel upload?',
  'home.cancelUploadMessage':
    'The upload is still in progress. If you cancel now, the partially uploaded file will be discarded and you will need to start over.',
  'home.keepUploading': 'Keep uploading',
  'home.collapseUpload': 'Minimize upload indicator',
  'home.expandUpload': 'Expand upload indicator',
  'home.deleted': 'Deleted "{name}"',
  'home.downloadStarting': 'Preparing download...',
  'home.downloadComplete': 'Download complete',
  'home.loadingFiles': 'Loading files...',
  'home.folderEmpty': 'This folder is empty',
  'home.uploadOrCreate': 'Upload files or create a new folder to get started',
  'home.noMatches': 'No files match your search',
  'home.tryDifferent': 'Try a different search term',
  'home.filterPlaceholder': 'Filter files...',
  'home.root': 'Root',
  'home.name': 'Name',
  'home.size': 'Size',
  'home.modified': 'Modified',
  'home.actions': 'Actions',
  'home.downloadZip': 'Download as ZIP',
  'home.download': 'Download',
  'home.downloadFailed': 'Download failed',
  'home.deleteConfirmTitle': 'Delete item?',
  'home.deleteConfirmMessage': 'Are you sure you want to delete "{name}"? This cannot be undone.',
  'home.deleteFolderNote': 'This folder and everything inside it will be permanently removed.',
  'home.totalSize': 'Total size: {size}',
  'home.file': 'file',
  'home.clearSearch': 'Clear search',

  // Summary
  'summary.title': 'Storage Summary',
  'summary.overview': 'Overview of your files and storage usage',
  'summary.loading': 'Loading summary...',
  'summary.loadFailed': 'Failed to load storage summary',
  'summary.totalStorage': 'Total Storage',
  'summary.files': 'Files',
  'summary.folders': 'Folders',
  'summary.fileTypes': 'File Types',
  'summary.storageByType': 'Storage by Type',
  'summary.noFilesYet': 'No files uploaded yet',
  'summary.typeImages': 'Images',
  'summary.typeVideo': 'Video',
  'summary.typeAudio': 'Audio',
  'summary.typeArchives': 'Archives',
  'summary.typeDocuments': 'Documents',
  'summary.typeOther': 'Other',

  // Chat / search
  'chat.title': 'Search',
  'chat.subtitle': 'Find files and folders across your storage',
  'chat.placeholder': 'Search by name or path...',
  'chat.results': '{n} result(s)',
  'chat.folder': 'Folder',
  'chat.noResults': 'No results found',
  'chat.tryDifferent': 'Try a different search term',
  'chat.searchFailed': 'Search failed',
  'chat.searchYourFiles': 'Search your files',
  'chat.typeQuery': 'Type a query above to find files and folders',

  // Preview
  'preview.loading': 'Loading preview...',
  'preview.couldNotOpen': 'Could not open this file in the viewer',
  'preview.noViewer': 'No in-app viewer for this file type',
  'preview.downloadInstead': 'Download instead',
  'preview.editMarkdown': 'Edit',
  'preview.saveFailed': 'Failed to save changes',
  'preview.saved': 'File saved',
  'preview.editing': 'Editing',
  'preview.loadFailed': 'Failed to load preview',
  'preview.truncated': '… preview truncated',
  'preview.ariaLabel': 'Preview {name}',
} as const;

export type MessageKey = keyof typeof en;

type Messages = Record<MessageKey, string>;

const ar: Messages = {
  'nav.myFiles': 'ملفاتي',
  'nav.summary': 'الملخص',
  'nav.activity': 'النشاط',
  'nav.upload': 'رفع',
  'nav.newFolder': 'مجلد جديد',
  'nav.cloudStorage': 'التخزين السحابي',
  'nav.language': 'اللغة',
  'nav.theme': 'المظهر',
  'nav.config': 'الإعدادات',
  'nav.shared': 'المشترك',
  'nav.admin': 'المستخدمون',
  'nav.signOut': 'تسجيل الخروج',

  'common.close': 'إغلاق',
  'common.cancel': 'إلغاء',
  'common.create': 'إنشاء',
  'common.save': 'حفظ',
  'common.download': 'تحميل',
  'common.preview': 'معاينة',
  'common.edit': 'تعديل',
  'common.retry': 'إعادة المحاولة',
  'common.delete': 'حذف',
  'common.openInNewTab': 'فتح في تبويب جديد',
  'common.back': 'رجوع',
  'common.dismiss': 'تجاهل',

  // Configuration panel
  'config.theme.system': 'النظام',
  'config.theme.dark': 'داكن',
  'config.theme.light': 'فاتح',
  'config.maxUploads': 'أقصى رفع متزامن',
  'config.hint': 'يحدد عدد الملفات التي تُرفع في وقت واحد (1–8).',

  'auth.connecting': 'جارٍ الاتصال...',
  'auth.unableToReach': 'تعذّر الوصول إلى الخادم',
  'auth.signIn': 'تسجيل الدخول',
  'auth.signInToAccess': 'سجّل الدخول للوصول إلى ملفاتك',
  'auth.username': 'اسم المستخدم',
  'auth.password': 'كلمة المرور',
  'auth.invalidCredentials': 'اسم المستخدم أو كلمة المرور غير صحيحة',
  'auth.createAccount': 'إنشاء حساب',
  'auth.noAccount': 'ليس لديك حساب؟',
  'auth.haveAccount': 'لديك حساب بالفعل؟',
  'auth.signUp': 'إنشاء حساب',
  'auth.signupDisabled': 'التسجيل معطّل',
  'auth.confirmPassword': 'تأكيد كلمة المرور',
  'auth.passwordsMismatch': 'كلمتا المرور غير متطابقتين',
  'auth.usernameHint': 'من 3 إلى 32 حرفًا: أحرف وأرقام ونقاط وشرطات وعلامات سفلية',
  'auth.passwordHint': '4 أحرف على الأقل',

  'register.title': 'إنشاء حساب',
  'register.subtitle': 'سجّل للبدء في تخزين ملفاتك',
  'register.submit': 'إنشاء حساب',
  'register.failed': 'تعذّر إنشاء الحساب',

  'admin.title': 'المستخدمون',
  'admin.subtitle': 'إدارة الحسابات واستخدام التخزين',
  'admin.loading': 'جارٍ تحميل المستخدمين...',
  'admin.loadFailed': 'فشل تحميل المستخدمين',
  'admin.username': 'اسم المستخدم',
  'admin.role': 'الدور',
  'admin.roleAdmin': 'مشرف',
  'admin.roleUser': 'مستخدم',
  'admin.files': 'الملفات',
  'admin.size': 'التخزين',
  'admin.createdAt': 'تاريخ الانضمام',
  'admin.empty': 'لا يوجد مستخدمون',

  'shared.readOnly': 'المجلد المشترك للقراءة فقط',

  'folder.created': 'تم إنشاء المجلد',
  'folder.nameRequired': 'اسم المجلد مطلوب',
  'folder.namePlaceholder': 'اسم المجلد',

  'home.items': '{n} عنصر',
  'home.refresh': 'تحديث',
  'home.upload': 'رفع',
  'home.uploading': 'جارٍ الرفع...',
  'home.uploaded': 'تم رفع {n} ملف بنجاح',
  'home.uploadCancelled': 'تم إلغاء الرفع',
  'home.cancelUpload': 'إلغاء الرفع',
  'home.cancelUploadTitle': 'إلغاء الرفع؟',
  'home.cancelUploadMessage':
    'لا يزال الرفع جارياً. إذا ألغيت الآن، سيتم تجاهل الملف المرفوع جزئياً وستحتاج إلى البدء من جديد.',
  'home.keepUploading': 'متابعة الرفع',
  'home.collapseUpload': 'تصغير مؤشر الرفع',
  'home.expandUpload': 'توسيع مؤشر الرفع',
  'home.deleted': 'تم حذف "{name}"',
  'home.downloadStarting': 'جاري تحضير التحميل...',
  'home.downloadComplete': 'اكتمل التحميل',
  'home.loadingFiles': 'جارٍ تحميل الملفات...',
  'home.folderEmpty': 'هذا المجلد فارغ',
  'home.uploadOrCreate': 'ارفع ملفات أو أنشئ مجلدًا جديدًا للبدء',
  'home.noMatches': 'لا توجد ملفات تطابق بحثك',
  'home.tryDifferent': 'جرّب مصطلح بحث مختلف',
  'home.filterPlaceholder': 'تصفية الملفات...',
  'home.root': 'الجذر',
  'home.name': 'الاسم',
  'home.size': 'الحجم',
  'home.modified': 'آخر تعديل',
  'home.actions': 'إجراءات',
  'home.downloadZip': 'تنزيل كملف مضغوط',
  'home.download': 'تنزيل',
  'home.downloadFailed': 'فشل التنزيل',
  'home.deleteConfirmTitle': 'حذف العنصر؟',
  'home.deleteConfirmMessage': 'هل أنت متأكد من حذف "{name}"؟ لا يمكن التراجع عن هذا الإجراء.',
  'home.deleteFolderNote': 'سيتم حذف هذا المجلد وكل ما بداخله نهائياً.',
  'home.totalSize': 'الحجم الإجمالي: {size}',
  'home.file': 'ملف',
  'home.clearSearch': 'مسح البحث',

  'summary.title': 'ملخص التخزين',
  'summary.overview': 'نظرة عامة على ملفاتك واستخدام التخزين',
  'summary.loading': 'جارٍ تحميل الملخص...',
  'summary.loadFailed': 'فشل تحميل ملخص التخزين',
  'summary.totalStorage': 'إجمالي التخزين',
  'summary.files': 'الملفات',
  'summary.folders': 'المجلدات',
  'summary.fileTypes': 'أنواع الملفات',
  'summary.storageByType': 'التخزين حسب النوع',
  'summary.noFilesYet': 'لا توجد ملفات مرفوعة بعد',
  'summary.typeImages': 'الصور',
  'summary.typeVideo': 'فيديو',
  'summary.typeAudio': 'صوت',
  'summary.typeArchives': 'أرشيفات',
  'summary.typeDocuments': 'مستندات',
  'summary.typeOther': 'أخرى',

  'chat.title': 'بحث',
  'chat.subtitle': 'ابحث في ملفاتك ومجلداتك عبر التخزين',
  'chat.placeholder': 'البحث بالاسم أو المسار...',
  'chat.results': '{n} نتيجة',
  'chat.folder': 'مجلد',
  'chat.noResults': 'لا توجد نتائج',
  'chat.tryDifferent': 'جرّب مصطلح بحث مختلف',
  'chat.searchFailed': 'فشل البحث',
  'chat.searchYourFiles': 'ابحث في ملفاتك',
  'chat.typeQuery': 'اكتب استعلامًا أعلاه للعثور على الملفات والمجلدات',

  'preview.loading': 'جارٍ تحميل المعاينة...',
  'preview.couldNotOpen': 'تعذّر فتح هذا الملف في العارض',
  'preview.noViewer': 'لا يوجد عارض مضمّن لهذا النوع من الملفات',
  'preview.downloadInstead': 'قم بالتنزيل بدلًا من ذلك',
  'preview.editMarkdown': 'تعديل',
  'preview.saveFailed': 'فشل حفظ التغييرات',
  'preview.saved': 'تم حفظ الملف',
  'preview.editing': 'جارٍ التعديل',
  'preview.loadFailed': 'فشل تحميل المعاينة',
  'preview.truncated': '… تم اقتطاع المعاينة',
  'preview.ariaLabel': 'معاينة {name}',
} ;

interface I18nContextValue {
  lang: Lang;
  dir: 'ltr' | 'rtl';
  setLang: (lang: Lang) => void;
  t: (key: MessageKey, vars?: Record<string, string | number>) => string;
}

const I18nContext = createContext<I18nContextValue | null>(null);

const STORAGE_KEY = 'fs_lang';

function detectLang(): Lang {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === 'en' || stored === 'ar') return stored;
    if (typeof navigator !== 'undefined') {
      const preferred = navigator.language?.toLowerCase() ?? '';
      if (preferred.startsWith('ar')) return 'ar';
    }
  } catch {
    /* storage unavailable */
  }
  return 'en';
}

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(detectLang);
  const dir: 'ltr' | 'rtl' = lang === 'ar' ? 'rtl' : 'ltr';

  useEffect(() => {
    const root = document.documentElement;
    root.lang = lang;
    root.dir = dir;
    try {
      localStorage.setItem(STORAGE_KEY, lang);
    } catch {
      /* ignore */
    }
  }, [lang, dir]);

  const setLang = useCallback((next: Lang) => setLangState(next), []);

  const t = useCallback(
    (key: MessageKey, vars?: Record<string, string | number>) => {
      const template: string = (lang === 'en' ? en : ar)[key];
      if (!vars) return template;
      return template.replace(/\{(\w+)\}/g, (_, name: string) =>
        String(vars[name] ?? `{${name}}`),
      );
    },
    [lang],
  );

  const value = useMemo(
    () => ({ lang, dir, setLang, t }),
    [lang, dir, setLang, t],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

// eslint-disable-next-line react-refresh/only-export-components
export function useI18n(): I18nContextValue {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error('useI18n must be used within <LanguageProvider>');
  return ctx;
}

// Convenience export for choosing the current font direction-aware classes.
export type { I18nContextValue };