-- migration up
ALTER TABLE ma_exam_sessions
  ADD COLUMN esess_flag TINYINT(1) DEFAULT NULL AFTER res_review;

UPDATE ma_exam_sessions
   SET esess_flag = (COALESCE(res_score_percentage, 0) >= 50)
 WHERE req_exam_type = 'GRADE' AND esess_status = 'COMPLETE';
