#!/usr/bin/env bash
# bin/lib/activate.sh — start server, verify health, restart nginx

run_activate() {
  phase_start "ACTIVATE"

  # Run existing post-deploy scripts (reuses current scripts)
  info "Running post-deploy scripts..."
  remote_exec "
    for i in $DEST_DIR/post-deploy/*.sh; do
      [ -f \"\$i\" ] && { echo \"Running \$i...\"; bash \"\$i\"; }
    done
  "

  if [ $? -ne 0 ]; then
    error "Post-deploy scripts failed"
    phase_end "ACTIVATE (with errors)"
    return 1
  fi
  
  info "Post-deploy scripts completed successfully"
  
  phase_end "ACTIVATE"
}
