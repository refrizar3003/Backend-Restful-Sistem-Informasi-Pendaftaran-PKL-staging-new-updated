package infrastructure

import "fmt"

func GetResetPasswordTemplate(otp string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Password Reset Code</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0;">
    <div style="width: 90%%; max-width: 600px; margin: 20px auto; border: 1px solid #ddd; border-radius: 8px; overflow: hidden;">
        <div style="padding: 20px;">
            <h2 style="color: #333; text-align: left;">Password Reset Code</h2>
            <p>Hello,</p>
            <p>We received a request to reset your password. Your password reset code is shown below:</p>
            <div style="background-color: #ffffff; border: 1px solid #f08080; border-radius: 8px; padding: 20px 30px; margin: 20px 0; text-align: center;">
                <strong style="font-size: 28px; color: #ec1e28; letter-spacing: 4px; display: block;">%s</strong>
            </div>
            <p style="margin-top: 20px;">
                This code will expire in <strong>10 minutes</strong>. If you did not request a password reset, please disregard this email.
            </p>
        </div>
        <div style="background-color: #f9f9f9; padding: 20px; text-align: center; border-top: 1px solid #ddd;">
            <p style="font-size: 12px; color: #888; margin: 0;">
                &copy; 2026 Telkom Witel Lampung. All rights reserved.
            </p>
        </div>
    </div>
</body>
</html>`, otp)
}

func GetVerificationEmailTemplate(verifyLink string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Verifikasi Akun</title>
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 0;">
    <div style="max-width: 600px; margin: 40px auto; background-color: #ffffff; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.1); overflow: hidden;">
        <div style="padding: 30px; color: #333333; line-height: 1.6;">
            <p>Halo,</p>
            <p>Berikut kami informasikan email <strong>Verifikasi Akun</strong> kamu. Silakan klik tombol di bawah ini untuk menyelesaikan proses registrasi akun.</p>
 
            <div style="text-align: center; margin-top: 20px;">
                <a href="%s"
                   style="display: inline-block; background-color: #e52d27; color: #ffffff; text-decoration: none; padding: 12px 25px; border-radius: 5px; font-weight: bold;">
                    Verifikasi Akun
                </a>
            </div>
 
            <p style="margin-top: 30px; font-size: 14px;">
                Tombol verifikasi hanya dapat diklik <strong>satu kali dan berlaku selama 24 jam.</strong>
                Jika masa berlaku habis, silakan lakukan pendaftaran atau permintaan verifikasi ulang.
            </p>
            <p style="font-size: 14px;">
                Terima kasih,<br>Administrator
            </p>
        </div>
        <div style="background-color: #f4f4f4; color: #777777; text-align: center; padding: 15px; font-size: 12px;">
            <p style="margin: 0;">&copy; 2026 Telkom Witel Lampung. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, verifyLink)
}

func GetApplicationAcceptedTemplate(fullname, unitName, mentorName, attendanceDate string) string {
    return fmt.Sprintf (`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Pengumuman Hasil Pendaftaran Praktik Kerja Lapangan Telkom Witel Lampung</title>
</head>
<body style="font-family: Arial, sans-serif; background-color:#f4f4f4; padding: 20px;">
    <div style="max-width:600px;margin:0 auto;background:#ffffff;border-radius:8px;overflow:hidden;">
        <div style="padding:24px;color:#202120;">
            <p>Halo <b>%s</b>,</p>
            <p>Selamat! Pengajuan Praktik Kerja Lapangan (PKL) kamu di Telkom Witel Lampung telah <b>diterima</b>.</p>
            <table style="width:100%%;border-collapse:collapse;margin:16px 0;">
                <tr><td style="padding:6px 0;color:#717E95;">Unit Penempatan</td><td style="padding:6px 0;font-weight:bold;">%s</td></tr>
                <tr><td style="padding:6px 0;color:#717E95;">Mentor</td><td style="padding:6px 0;font-weight:bold;">%s</td></tr>
                <tr><td style="padding:6px 0;color:#717E95;">Tanggal Mulai</td><td style="padding:6px 0;font-weight:bold;">%s</td></tr>
            </table>
            <p>Silakan login ke akun kamu untuk melihat surat penerimaan resmi dan detail lebih lanjut.</p>
            <p style="font-size:14px;margin-top:24px;">Terima kasih,<br>Administrator</p>
        </div>
        <div style="background-color:#f4f4f4;color:#777777;text-align:center;padding:15px;font-size:12px;">
            <p style="margin:0;">&copy; 2026 Telkom Witel Lampung. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, fullname, unitName, mentorName, attendanceDate)
}

func GetApplicationRejectedTemplate(fullname, adminNotes string) string {
    if adminNotes == "" {
        adminNotes = "Tidak ada catatan tambahan dari admin."
    }
    return fmt.Sprintf (`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Pengumuman Hasil Pendaftaran Praktik Kerja Lapangan Telkom Witel Lampung</title>
</head>
<body style="font-family: Arial, sans-serif; background-color:#f4f4f4; padding: 20px;">
    <div style="max-width:600px;margin:0 auto;background:#ffffff;border-radius:8px;overflow:hidden;">
        <div style="padding:24px;color:#202120;">
            <p>Halo <b>%s</b>,</p>
            <p>Mohon maaf, pengajuan Praktik Kerja Lapangan (PKL) kamu di Telkom Witel Lampung <b>belum dapat kami terima</b> pada periode ini.</p>
            <div style="background:#f9f9f9;border-left:4px solid #EC1E28;padding:12px 16px;margin:16px 0;">
                <p style="margin:0;"><b>Catatan Admin:</b> %s</p>
            </div>
            <p>Kamu tetap dapat mendaftar kembali pada periode berikutnya. Terima kasih atas minatmu.</p>
            <p style="font-size:14px;margin-top:24px;">Terima kasih,<br>Administrator</p>
        </div>
        <div style="background-color:#f4f4f4;color:#777777;text-align:center;padding:15px;font-size:12px;">
            <p style="margin:0;">&copy; 2026 Telkom Witel Lampung. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, fullname, adminNotes)
}

func GetLoginOTPTemplate(otp string) string {
    return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Password Reset Code</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0;">
    <div style="width: 90%%; max-width: 600px; margin: 20px auto; border: 1px solid #ddd; border-radius: 8px; overflow: hidden;">
        <div style="padding: 20px;">
            <h2 style="color: #333; text-align: left;">Kode OTP Login</h2>
            <p>Hai,</p>
            <p>Kami menerima permintaan login ke akun Kamu. Kode OTP Kamu:</p>
            <div style="background-color: #ffffff; border: 1px solid #f08080; border-radius: 8px; padding: 20px 30px; margin: 20px 0; text-align: center;">
                <strong style="font-size: 28px; color: #ec1e28; letter-spacing: 4px; display: block;">%s</strong>
            </div>
            <p style="margin-top: 20px;">
                Kode ini akan kedaluwarsa dalam <strong>5 menit</strong>. Jika Kamu tidak melakukan permintaan login, abaikan email ini.
            </p>
        </div>
        <div style="background-color: #f9f9f9; padding: 20px; text-align: center; border-top: 1px solid #ddd;">
            <p style="font-size: 12px; color: #888; margin: 0;">
                &copy; 2026 Telkom Witel Lampung. All rights reserved.
            </p>
        </div>
    </div>
</body>
</html>`, otp)
}