-- migration up
-- Seed the ma_seqs registry for the new junction aggregate so commands
-- can call repos.Seq.Next(ctx, seq.NameClassroomProgram). Boot-time
-- auto-migrate is disabled (see internal/bootstrap/app.go), so apply
-- this file manually alongside 018/019.
INSERT INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('classroom_program', 0, 'CP', 8);  -- classroom_program_id: CP00000001...
