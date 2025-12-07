-- migration up
CREATE TABLE chapters (
    `id` CHAR(36) NOT NULL,
    `grade_id` CHAR(36) NOT NULL,
    `semester_id` CHAR(36) NOT NULL,
    `chapter_number` INT NOT NULL,
    `title` VARCHAR(200) NOT NULL,
    `description` TEXT DEFAULT NULL,
    `languague` VARCHAR(10) NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- -- Grade 1 - Midterm 1 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 1, 'Numbers 1-10', 'Introduction to counting and recognizing numbers from 1 to 10', 'EN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 2, 'Basic Addition', 'Understanding addition within 10', 'EN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 3, 'Shapes and Colors', 'Identifying basic shapes and colors', 'EN');

-- -- Grade 1 - Midterm 1 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 1, 'Các số từ 1 đến 10', 'Giới thiệu về đếm và nhận biết các số từ 1 đến 10', 'VN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 2, 'Phép cộng cơ bản', 'Hiểu về phép cộng trong phạm vi 10', 'VN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 3, 'Hình dạng và Màu sắc', 'Nhận biết các hình cơ bản và màu sắc', 'VN');

-- -- Grade 1 - End of Term 1 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 1, 'Patterns and Sequences', 'Creating and extending simple patterns', 'EN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 2, 'Measurement Basics', 'Comparing length, weight, and capacity', 'EN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 3, 'Time Basics', 'Understanding hours and half hours', 'EN');

-- -- Grade 1 - End of Term 1 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 1, 'Mẫu và Dãy số', 'Tạo và mở rộng các mẫu đơn giản', 'VN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 2, 'Đo lường cơ bản', 'So sánh chiều dài, khối lượng và dung tích', 'VN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 3, 'Thời gian cơ bản', 'Hiểu về giờ đúng và nửa giờ', 'VN');

-- -- Grade 1 - Midterm 2 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 1, 'Numbers 11-20', 'Extending number knowledge to twenty', 'EN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 2, 'Basic Subtraction', 'Understanding subtraction within 10', 'EN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 3, 'Position and Direction', 'Understanding spatial relationships', 'EN');

-- -- Grade 1 - Midterm 2 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 1, 'Các số từ 11 đến 20', 'Mở rộng kiến thức về số đến hai mươi', 'VN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 2, 'Phép trừ cơ bản', 'Hiểu về phép trừ trong phạm vi 10', 'VN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 3, 'Vị trí và Hướng', 'Hiểu về mối quan hệ không gian', 'VN');

-- -- Grade 1 - End of Term 2 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 1, 'Numbers to 50', 'Counting and understanding numbers up to 50', 'EN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 2, 'Addition and Subtraction Review', 'Practicing addition and subtraction skills', 'EN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 3, 'Money Introduction', 'Identifying coins and their values', 'EN');

-- -- Grade 1 - End of Term 2 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 1, 'Các số đến 50', 'Đếm và hiểu về các số đến 50', 'VN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 2, 'Ôn tập cộng và trừ', 'Luyện tập kỹ năng cộng và trừ', 'VN'),
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 3, 'Giới thiệu tiền', 'Nhận biết tiền và giá trị của chúng', 'VN');

-- -- Grade 2 - Midterm 1 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 1, 'Numbers to 100', 'Understanding place value and numbers up to 100', 'EN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 2, 'Addition within 100', 'Two-digit addition with and without regrouping', 'EN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 3, 'Time and Money', 'Reading clocks and counting money', 'EN');

-- -- Grade 2 - Midterm 1 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 1, 'Các số đến 100', 'Hiểu giá trị chữ số và các số đến 100', 'VN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 2, 'Phép cộng trong phạm vi 100', 'Cộng hai chữ số có và không có nhớ', 'VN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 3, 'Thời gian và Tiền', 'Đọc đồng hồ và đếm tiền', 'VN');

-- -- Grade 2 - End of Term 1 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 1, 'Geometry Basics', 'Identifying and drawing 2D shapes', 'EN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 2, 'Measurement Units', 'Understanding centimeters, meters, grams, and kilograms', 'EN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 3, 'Word Problems', 'Solving multi-step word problems', 'EN');

-- -- Grade 2 - End of Term 1 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 1, 'Hình học cơ bản', 'Nhận biết và vẽ các hình 2D', 'VN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 2, 'Đơn vị đo lường', 'Hiểu về centimet, mét, gram và kilogram', 'VN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 3, 'Bài toán có lời văn', 'Giải bài toán nhiều bước', 'VN');

-- -- Grade 2 - Midterm 2 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 1, 'Subtraction within 100', 'Two-digit subtraction with and without regrouping', 'EN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 2, 'Introduction to Multiplication', 'Understanding multiplication as repeated addition', 'EN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 3, 'Data and Graphs', 'Reading and creating simple graphs', 'EN');

-- -- Grade 2 - Midterm 2 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 1, 'Phép trừ trong phạm vi 100', 'Trừ hai chữ số có và không có mượn', 'VN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 2, 'Giới thiệu phép nhân', 'Hiểu phép nhân như phép cộng lặp lại', 'VN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 3, 'Dữ liệu và Biểu đồ', 'Đọc và tạo biểu đồ đơn giản', 'VN');

-- -- Grade 2 - End of Term 2 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 1, 'Fractions Introduction', 'Understanding halves and quarters', 'EN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 2, 'Times Tables Practice', 'Multiplication facts for 2, 5, and 10', 'EN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 3, 'Line Graphs', 'Reading and creating line graphs', 'EN');

-- -- Grade 2 - End of Term 2 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 1, 'Giới thiệu phân số', 'Hiểu về một nửa và một phần tư', 'VN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 2, 'Luyện tập bảng nhân', 'Bảng nhân 2, 5 và 10', 'VN'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 3, 'Biểu đồ đường', 'Đọc và tạo biểu đồ đường', 'VN');

-- -- Grade 3 - Midterm 1 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 1, 'Numbers to 1000', 'Place value, rounding, and comparing three-digit numbers', 'EN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 2, 'Multiplication Facts', 'Mastering multiplication tables 0-10', 'EN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 3, 'Fractions', 'Understanding parts of a whole and simple fractions', 'EN');

-- -- Grade 3 - Midterm 1 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 1, 'Các số đến 1000', 'Giá trị chữ số, làm tròn và so sánh số có ba chữ số', 'VN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 2, 'Bảng nhân', 'Thành thạo bảng nhân từ 0-10', 'VN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 3, 'Phân số', 'Hiểu phần của tổng thể và phân số đơn giản', 'VN');

-- -- Grade 3 - End of Term 1 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 1, 'Adding and Subtracting within 1000', 'Multi-digit addition and subtraction', 'EN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 2, 'Equivalent Fractions', 'Finding fractions with same value', 'EN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 3, 'Time and Calendar', 'Reading time and understanding calendars', 'EN');

-- -- Grade 3 - End of Term 1 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 1, 'Cộng và trừ trong phạm vi 1000', 'Cộng và trừ số có nhiều chữ số', 'VN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 2, 'Phân số bằng nhau', 'Tìm phân số có giá trị bằng nhau', 'VN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 3, 'Thời gian và Lịch', 'Đọc thời gian và hiểu lịch', 'VN');

-- -- Grade 3 - Midterm 2 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 1, 'Division Basics', 'Understanding division as sharing and grouping', 'EN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 2, 'Geometry', 'Exploring 2D and 3D shapes, perimeter, and area', 'EN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 3, 'Measurement', 'Length, mass, volume, and temperature', 'EN');

-- -- Grade 3 - Midterm 2 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 1, 'Phép chia cơ bản', 'Hiểu phép chia như chia sẻ và nhóm', 'VN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 2, 'Hình học', 'Khám phá hình 2D và 3D, chu vi và diện tích', 'VN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 3, 'Đo lường', 'Chiều dài, khối lượng, thể tích và nhiệt độ', 'VN');

-- -- Grade 3 - End of Term 2 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 1, 'Decimals Introduction', 'Understanding tenths and hundredths', 'EN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 2, 'Multiplication and Division Word Problems', 'Solving real-world multiplication and division problems', 'EN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 3, 'Angles and Lines', 'Understanding angles, parallel and perpendicular lines', 'EN');

-- -- Grade 3 - End of Term 2 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 1, 'Giới thiệu số thập phân', 'Hiểu về phần mười và phần trăm', 'VN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 2, 'Bài toán nhân và chia', 'Giải bài toán thực tế về nhân và chia', 'VN'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 3, 'Góc và Đường thẳng', 'Hiểu về góc, đường song song và vuông góc', 'VN');

-- -- Grade 4 - Midterm 1 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 1, 'Large Numbers', 'Reading, writing, and comparing numbers to millions', 'EN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 2, 'Multi-Digit Multiplication', 'Multiplying multi-digit numbers', 'EN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 3, 'Factors and Multiples', 'Finding factors, multiples, and prime numbers', 'EN');

-- -- Grade 4 - Midterm 1 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 1, 'Số lớn', 'Đọc, viết và so sánh số đến hàng triệu', 'VN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 2, 'Phép nhân nhiều chữ số', 'Nhân các số có nhiều chữ số', 'VN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 3, 'Ước và Bội', 'Tìm ước, bội và số nguyên tố', 'VN');

-- -- Grade 4 - End of Term 1 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 1, 'Fractions and Decimals', 'Converting between fractions and decimals', 'EN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 2, 'Adding and Subtracting Fractions', 'Operations with like denominators', 'EN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 3, 'Lines and Angles', 'Types of angles and measuring with protractor', 'EN');

-- -- Grade 4 - End of Term 1 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 1, 'Phân số và Số thập phân', 'Chuyển đổi giữa phân số và số thập phân', 'VN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 2, 'Cộng và trừ phân số', 'Phép tính với mẫu số giống nhau', 'VN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 3, 'Đường thẳng và Góc', 'Các loại góc và đo bằng thước đo góc', 'VN');

-- -- Grade 4 - Midterm 2 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 1, 'Multi-Digit Division', 'Dividing multi-digit numbers', 'EN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 2, 'Area and Perimeter', 'Calculating area and perimeter of rectangles', 'EN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 3, 'Symmetry', 'Lines of symmetry and rotational symmetry', 'EN');

-- -- Grade 4 - Midterm 2 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 1, 'Phép chia nhiều chữ số', 'Chia các số có nhiều chữ số', 'VN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 2, 'Diện tích và Chu vi', 'Tính diện tích và chu vi hình chữ nhật', 'VN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 3, 'Đối xứng', 'Trục đối xứng và đối xứng quay', 'VN');

-- -- Grade 4 - End of Term 2 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 1, 'Decimal Operations', 'Adding and subtracting decimals', 'EN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 2, 'Mixed Numbers', 'Converting between improper fractions and mixed numbers', 'EN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 3, 'Data Analysis', 'Interpreting tables, charts, and graphs', 'EN');

-- -- Grade 4 - End of Term 2 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 1, 'Phép tính số thập phân', 'Cộng và trừ số thập phân', 'VN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 2, 'Số hỗn hợp', 'Chuyển đổi giữa phân số không thực sự và số hỗn hợp', 'VN'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 3, 'Phân tích dữ liệu', 'Giải thích bảng, biểu đồ', 'VN');

-- -- Grade 5 - Midterm 1 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 1, 'Place Value to Billions', 'Understanding large numbers to billions', 'EN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 2, 'Operations with Decimals', 'Multiplying and dividing decimals', 'EN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 3, 'Powers and Exponents', 'Understanding powers of 10 and exponents', 'EN');

-- -- Grade 5 - Midterm 1 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 1, 'Giá trị chữ số đến hàng tỷ', 'Hiểu số lớn đến hàng tỷ', 'VN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 2, 'Phép tính với số thập phân', 'Nhân và chia số thập phân', 'VN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 3, 'Lũy thừa và Số mũ', 'Hiểu lũy thừa của 10 và số mũ', 'VN');

-- -- Grade 5 - End of Term 1 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 1, 'Operations with Fractions', 'Adding, subtracting, multiplying fractions', 'EN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 2, 'Volume', 'Finding volume of rectangular prisms', 'EN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 3, 'Coordinate Plane', 'Plotting points on coordinate grids', 'EN');

-- -- Grade 5 - End of Term 1 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 1, 'Phép tính với phân số', 'Cộng, trừ, nhân phân số', 'VN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 2, 'Thể tích', 'Tìm thể tích hình hộp chữ nhật', 'VN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '2c0a1e3b-3g4h-6f5d-0a2c-9e8f7g6h5a13', 3, 'Mặt phẳng tọa độ', 'Vẽ điểm trên lưới tọa độ', 'VN');

-- -- Grade 5 - Midterm 2 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 1, 'Ratios and Proportions', 'Understanding ratios and proportional relationships', 'EN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 2, 'Percentages', 'Converting between fractions, decimals, and percentages', 'EN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 3, 'Algebraic Expressions', 'Introduction to variables and simple expressions', 'EN');

-- -- Grade 5 - Midterm 2 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 1, 'Tỷ lệ và Tỷ số', 'Hiểu tỷ lệ và mối quan hệ tỷ lệ thuận', 'VN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 2, 'Phần trăm', 'Chuyển đổi giữa phân số, số thập phân và phần trăm', 'VN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 3, 'Biểu thức đại số', 'Giới thiệu biến và biểu thức đơn giản', 'VN');

-- -- Grade 5 - End of Term 2 - English
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 1, 'Statistics and Probability', 'Mean, median, mode, and basic probability', 'EN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 2, 'Transformations', 'Translation, reflection, and rotation', 'EN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 3, 'Problem Solving Review', 'Multi-step problems and real-world applications', 'EN');

-- -- Grade 5 - End of Term 2 - Vietnamese
-- INSERT INTO chapters (id, grade_id, semester_id, chapter_number, title, description, languague) VALUES
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 1, 'Thống kê và Xác suất', 'Trung bình, trung vị, yếu vị và xác suất cơ bản', 'VN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 2, 'Phép biến hình', 'Tịnh tiến, đối xứng và quay', 'VN'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', '3d1b2f4c-4h5i-7g6e-1b3d-0f9g8h7i6b24', 3, 'Ôn tập giải toán', 'Bài toán nhiều bước và ứng dụng thực tế', 'VN');
